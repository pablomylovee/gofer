/* This is the storage system. This handles reading,
   uploading, removing, moving, and renaming. Yet to
   include support for writing. */

package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type File struct {
	Name    string `json:"name"`
	RelPath string `json:"path"`
	ModTime int64  `json:"mod-time"`
	Size    int64  `json:"size"`
	File    bool   `json:"file"`
	Type    string `json:"type"`
}

// Is there enough space? // Is there enough space for others once I add... ?
func checkSize(user string, with ...int64) (bool, error) {
	var size int64;
	var files, err = os.ReadDir(filepath.Join(settings.Storage.Path, user));
	if err != nil {
		return false, err;
	}

	for _, file := range files {
		var info, err1 = file.Info();
		if err1 != nil {
			return false, err1;
		}
		size += info.Size();
	}

	if size >= settings.Storage.MaxSize {
		return false, nil;
	} else if len(with) == 1 {
		if size + with[0] > settings.Storage.MaxSize {
			return false, nil;
		}
	}

	return true, nil;
}

// GoFiber handler to list files in a directory
func http_dir(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}

	var path string = filepath.Join(settings.Storage.Path, username, filepath.Clean("/"+c.Params("*")));
	if s, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return c.SendStatus(fiber.StatusNotFound);
		}
	} else if s.Mode().IsRegular() {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}
	if content, found, err := getCache(username, c.Params("*")); err == nil && found {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8);
		return c.Send(content.([]byte));
	}

	var files, err = os.ReadDir(path);
	if err != nil {
		return c.Status(500).SendString(err.Error());
	}

	var body []File;
	for _, file := range files {
		if info, err1 := file.Info(); err1 != nil {
			return c.Status(500).SendString(err1.Error());
		} else {
			var f File = File{
				Name: info.Name(),
				RelPath: filepath.Clean(filepath.Join(c.Params("*"), info.Name())),
				ModTime: info.ModTime().Unix(),
				Size: info.Size(),
				File: file.Type().IsRegular(),
			}
			if f.File {
				f.Type = mime.TypeByExtension(filepath.Ext(f.Name));
			} else {
				f.Type = "directory";
			}
			if f.Type == "" {
				f.Type = "application/octet-stream";
			}
			body = append(body, f);
		}
	}

	var bbody, err1 = json.Marshal(body);
	if err1 != nil {
		return c.Status(500).SendString(err1.Error());
	}

	if _, err := setCache(username, c.Params("*"), bbody); err != nil {
		return c.Status(500).SendString(err.Error());
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8);
	return c.Send(bbody);
}

// GoFiber handler to send files to the user
func http_file(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}

	if stat, err := os.Stat(filepath.Join(settings.Storage.Path, username, c.Params("*")));
	err == nil && stat.IsDir() {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}
	return c.SendFile(filepath.Join(settings.Storage.Path, username, c.Params("*")));
}

// GoFiber handler for file deletion
func http_rm(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}
	if c.Query("path") == "" {
		return c.SendStatus(fiber.StatusBadRequest);
	}
	var relPath string = filepath.Clean(c.Query("path"));
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}

	var userDir string = filepath.Join(settings.Storage.Path, username);
	var fullPath string = filepath.Join(userDir, relPath);
	
	var udAbs, err = filepath.Abs(userDir);
	var fpAbs, err1 = filepath.Abs(fullPath);
	if err != nil || err1 != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	if !strings.HasPrefix(fpAbs, udAbs+string(os.PathSeparator)) {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}

	if err := os.RemoveAll(fpAbs); err != nil {
		if os.IsNotExist(err) {
			return c.SendStatus(fiber.StatusNotFound);
		}
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	return c.SendStatus(fiber.StatusOK);
}

func http_mv(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}

	if c.Query("fr") == "" || c.Query("to") == "" {
		return c.SendStatus(fiber.StatusBadRequest);
	}

	var frRelPath string = filepath.Clean(c.Query("fr"));
	if frRelPath == ".." || strings.HasPrefix(frRelPath, ".."+string(os.PathSeparator)) {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}
	var toRelPath string = filepath.Clean(c.Query("to"));
	if toRelPath == ".." || strings.HasPrefix(toRelPath, ".."+string(os.PathSeparator)) {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}

	var userDir string = filepath.Join(settings.Storage.Path, username);
	var frFullPath, toFullPath string = filepath.Join(userDir, frRelPath), filepath.Join(userDir, toRelPath);

	var udAbs string;
	if x, err := filepath.Abs(userDir); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else {
		udAbs = x;
	}
	var frFpAbs, toFpAbs string;
	if ffa, err := filepath.Abs(frFullPath); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else {
		frFpAbs = ffa;
	}
	if tfa, err := filepath.Abs(toFullPath); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else {
		toFpAbs = tfa;
	}

	if !strings.HasPrefix(frFpAbs, udAbs+string(os.PathSeparator)) ||
	!strings.HasPrefix(toFpAbs, udAbs+string(os.PathSeparator)) {
		return c.SendStatus(fiber.StatusNotAcceptable);
	}

	if err := os.Rename(frFpAbs, toFpAbs); err != nil {
		if os.IsNotExist(err) {
			return c.SendStatus(fiber.StatusNotFound);
		}
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	return c.SendStatus(fiber.StatusOK);
}

func http_new(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}

	var size int64;
	if x, err := strconv.Atoi(c.Query("s")); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else {
		size = int64(x);
	}
	if yes, err := checkSize(username, size); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else if !yes {
		return c.SendStatus(fiber.StatusServiceUnavailable);
	}
	
	var code int = rand.IntN(1000000) + 100000;

	var fragPath, err = here("fragments", fmt.Sprintf("%s-%d", username, code));
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	if err := os.MkdirAll(fragPath, 0755); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	return c.Status(fiber.StatusCreated).SendString(strconv.Itoa(code));
}

func http_add(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}

	var path, err2 = here("fragments", fmt.Sprintf("%s-%s", username, c.Query("id")), c.Query("n")+".frag");
	if err2 != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	var fragment, err3 = os.Create(path);
	if err3 != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	defer fragment.Close();

	if _, err := fragment.Write(c.Body()); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	return c.SendStatus(fiber.StatusCreated);
}

func http_create(c fiber.Ctx) error {
	var username, a = authenticate(c);
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized);
	}

	var size int64;
	if x, err := strconv.Atoi(c.Query("s")); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else {
		size = int64(x);
	}
	if yes, err := checkSize(username, size); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	} else if !yes {
		return c.SendStatus(fiber.StatusServiceUnavailable);
	}
	
	var fragPath, err = here("fragments", fmt.Sprintf("%s-%s", username, c.Query("id")));
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	var frags, err1 = os.ReadDir(fragPath);
	if err1 != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	var file, err2 = os.Create(filepath.Join(settings.Storage.Path, username, c.Query("name")));
	if err2 != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	defer file.Close();

	sort.Slice(frags, func(i int, j int) bool {
		var iname, jname string = frags[i].Name(), frags[j].Name();
		var iorder, _ = strconv.Atoi(strings.TrimSuffix(iname, ".frag"));
		var jorder, _ = strconv.Atoi(strings.TrimSuffix(jname, ".frag"));

		return iorder < jorder;
	});

	for _, frag := range frags {
		var content, err3 = os.ReadFile(filepath.Join(fragPath, frag.Name()));
		if err3 != nil {
			return c.SendStatus(fiber.StatusInternalServerError);
		}
		file.Write(content);
	}

	if err := os.RemoveAll(fragPath); err != nil {
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	return c.SendStatus(fiber.StatusOK);
}

