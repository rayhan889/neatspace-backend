package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

func Error(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	// Return JSON response using BaseResponse format
	return c.Status(code).JSON(apputils.ErrorResponse(code, err.Error(), ""))
}
