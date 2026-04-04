/* Some other functions for other functions. */

package main

import (
	"path/filepath"
	"slices"

	"github.com/gofiber/fiber/v3"
)

// "Join this with the executable's location."
func here(p ...string) (string, error) {
	var h, err = filepath.Abs(".")
	if err != nil {
		return "", err
	}
	return filepath.Join(h, filepath.Join(p...)), nil
}

// "Are they allowed?"
func allowed(c fiber.Ctx) error {
	if len(settings.Whitelist) > 0 {
		if slices.Contains(settings.Whitelist, c.IP()) {
			return c.Next();
		}
		return c.SendStatus(fiber.StatusForbidden);
	}
	if slices.Contains(settings.Blacklist, c.IP()) {
		return c.SendStatus(fiber.StatusForbidden);
	}
	return c.Next();
}

