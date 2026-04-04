/** These are functions to read and write to the configuration file.
    ATM there's only a function to read. **/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type User struct {
	Name string `json:"name"`
	Pass string `json:"pass"`
	Sudo bool   `json:"sudo"`
}

type Settings struct {
	Port      int      `json:"port"`
	Users     []User   `json:"users"`
	Blacklist []string `json:"blacklist"`
	Whitelist []string `json:"whitelist"`
	GUI       struct {
		Use    bool   `json:"use"`
		Source string `json:"src"` // static files or special URL (f/url resp.)
		To     string `json:"to"`  // path to static files/URL
	} `json:"gui"`
	Cache     struct{
		LowEffort    bool    `json:"low-effort"`
		Memory       bool    `json:"memory"`
		MaxSize      uint    `json:"max-size"`
		Path         string  `json:"path"`
		MaxUnusedAge float32 `json:"max-unused-age"`
	} `json:"cache"`
	Storage   struct{
		Path    string `json:"path"`
		MaxSize int64   `json:"max-size"`
	} `json:"storage"`
}

// "Give me the settings."
func getSettings() (Settings, error) {
	var toReturn Settings;
	var path, err2 = here("config.json");
	if err2 != nil {
		return Settings{}, err2;
	}
	var config, err = os.ReadFile(path);
	if err != nil {
		return Settings{}, err
	}

	var err1 error = json.Unmarshal(config, &toReturn);
	if err1 != nil {
		var syntaxerr *json.SyntaxError;
		var typerr *json.UnmarshalTypeError;

		if errors.As(err1, &syntaxerr) || errors.As(err1, &typerr) {
			fmt.Println("error: bad configuration");
		}
		return Settings{}, err1;
	}

	return toReturn, nil
}

// "Is the configuration proper?"
func configCheck() (bool, string, error) {
	var path, err3 = here("config.json");
	if err3 != nil {
		return false, "", err3;
	}
	// "Does the file exist?"
	var _, err2 = os.Stat(path);
	if err2 != nil {
		if errors.Is(err2, fs.ErrNotExist) {
			return false, "no configuration", nil;
		}

		return false, "", err2;
	}
	var config, err = os.ReadFile(path);
	if err != nil {
		return false, "", err;
	}
	var s Settings;
	var err1 error = json.Unmarshal(config, &s);
	// "Is this JSON syntax correct?"
	if err1 != nil {
		var syntaxerr *json.SyntaxError;
		var typerr *json.UnmarshalTypeError;

		if errors.As(err1, &syntaxerr) || errors.As(err1, &typerr) {
			return false, "bad configuration", nil;
		}
		return false, "", err1;
	}

	// "Are there clones of users?"
	if len(s.Users) < 1 {
		return false, "no users", nil;
	}
	for _, user := range s.Users {
		var timesAppeared int = 0;
		for _, u := range s.Users {
			if user.Name == u.Name {
				timesAppeared += 1;
			}
		}
		if timesAppeared > 1 {
			return false, "users with the same name", nil;
		}
	}

	return true, "", nil
}

