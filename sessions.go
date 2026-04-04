/** This is mainly to hide usernames and passwords during commands.
    We use the username and password in the body as JSON only when
	creating/revoking a session, so that we may support hiding the
	username and password as much as we can. I haven't any option
	that I know of besides the JSON form, so that's how I'll do it
	for the creation of a session and the nullification of a session
	(requiring sudo user). **/

package main

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Every session created.
var sessions sync.Map

// "Do they have a session? If so, what's their name?"
func authenticate(c fiber.Ctx) (string, bool) {
	if user, ok := sessions.Load(c.Cookies("session-key")); ok {
		return user.(string), ok;
	}
	if after, ok := strings.CutPrefix(c.Get("Authorization"), "Bearer "); ok {
		if user, ok1 := sessions.Load(after); ok1 {
			return user.(string), true;
		}
	}
	return "", false;
}

// GoFiber handler for session creation
func http_createSession(c fiber.Ctx) error {
	// Authenticating - using username and password
	var form struct {
		Username string
		Password string
	}
	if err1 := json.Unmarshal(c.Body(), &form); err1 != nil {
		var syntaxError *json.SyntaxError
		var typeErr *json.UnmarshalTypeError

		if errors.As(err1, &syntaxError) || errors.As(err1, &typeErr) {
			return c.Status(fiber.StatusBadRequest).SendString("Bad login form")
		} else {
			return c.Status(fiber.StatusInternalServerError).SendString("Could not parse login form")
		}
	}
	var a bool = false
	for _, user := range settings.Users {
		if user.Name == form.Username && user.Pass == form.Password {
			a = true
			break
		}
	}
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Creating session
	var id string = uuid.NewString()
	for { // yes I'm this paranoid
		if _, ok := sessions.Load(id); !ok {
			break
		}
		id = uuid.NewString()
	}

	sessions.Store(id, form.Username)

	// Allowing usage
	c.ClearCookie("session-key")
	c.Cookie(&fiber.Cookie{
		HTTPOnly: true,
		Name:     "session-key",
		Value:    id,
	})

	return c.Status(fiber.StatusOK).SendString("Successfully created ID!")
}

// GoFiber handler for nullification of every session made by a user
func http_nullifySessions(c fiber.Ctx) error {
	// Authenticating - using username and password
	var form struct {
		Username string
		Password string
	}
	if err1 := json.Unmarshal(c.Body(), &form); err1 != nil {
		var syntaxError *json.SyntaxError
		var typeErr *json.UnmarshalTypeError

		if errors.As(err1, &syntaxError) || errors.As(err1, &typeErr) {
			return c.Status(fiber.StatusBadRequest).SendString("Bad login form")
		} else {
			return c.Status(fiber.StatusInternalServerError).SendString("Could not parse login form")
		}
	}

	var a bool = false
	var user User
	for _, u := range settings.Users {
		if u.Name == form.Username && u.Pass == form.Password {
			a = true
			user = u
			break
		}
	}
	if !a {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	if !user.Sudo {
		sessions.Range(func(key, value any) bool {
			if value.(string) == form.Username {
				sessions.Delete(key)
			}
			return true
		})
		c.ClearCookie("session-key")
		return c.Status(fiber.StatusOK).SendString("Successfully annuled every session.")
	}
	if c.Query("user") == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing user.")
	}
	sessions.Range(func(key, value any) bool {
		if value.(string) == c.Query("user") {
			sessions.Delete(key)
		}
		return true
	})
	return c.Status(fiber.StatusOK).SendString("sudo: Successfully annuled every session. (" + c.Query("user") + ")")
}
