package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cache"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/static"
	_ "github.com/mattn/go-sqlite3"
)

// Global, general settings
var settings Settings;

func main() {
	var valid, why, err = configCheck()
	if err != nil {
		panic(err);
	}
	if !valid && why != "" {
		fmt.Printf("error: %s\n", why);
		os.Exit(1);
	}
	if s, e := getSettings(); e == nil {
		settings = s;
	}
	if err1 := initCache(); err1 != nil {
		panic(err1);
	}
	for _, user := range settings.Users {
		if err1 := os.MkdirAll(filepath.Join(settings.Storage.Path, user.Name), 0755); err1 != nil {
			panic(err1);
		}
	}
	os.Mkdir("fragments", 0755);

	var app *fiber.App = fiber.New(fiber.Config{AppName: "gofer"})
	app.Use(favicon.New());
	app.Use(allowed);
	if settings.Cache.LowEffort {
		app.Use(cache.New(cache.Config{
			MaxBytes: settings.Cache.MaxSize,
		}));
	}
	app.Use(compress.New(compress.Config{
		Next: func(c fiber.Ctx) bool {
			return c.Query("compress", "no") != "yes";
		},
		Level: compress.LevelBestCompression,
	}));

	app.Post("/session/create", http_createSession);
	app.Post("/session/nullify", http_nullifySessions);

	app.Get("/dir/*", http_dir);
	app.Get("/file/*", http_file);
	app.Get("/rm", http_rm);
	app.Get("/mv", http_mv);
	app.Get("/create", http_create);
	app.Get("/new", http_new);
	app.Post("/add", http_add);

	if settings.GUI.Use {
		switch settings.GUI.Source {
			case "f": app.Get("/gui/*", static.New(settings.GUI.To));
			case "url":
				app.Get("/gui", func (c fiber.Ctx) error {
					return c.Redirect().To(settings.GUI.To);
				});
		}
	}

	app.Listen(fmt.Sprintf(":%d", settings.Port), fiber.ListenConfig{
		DisableStartupMessage: true,
	});
}
