package middlewares

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rayhan889/neatspace/internal/application/constants"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

var (
	validate = validator.New()
)

// ValidateRequestJSON is a middleware that validates the JSON request body against the struct T.
func ValidateRequestJSON[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		obj := new(T)

		// Parse the request body into the struct
		if err := c.BodyParser(obj); err != nil {
			log.Printf("Validation JSON error %v", err)

			if validatorErrors, isNoValid := err.(validator.ValidationErrors); isNoValid {
				errors := mappingErrorMessage(validatorErrors)

				return c.Status(fiber.StatusBadRequest).JSON(apputils.ErrorValidationResponse(fiber.StatusBadRequest, errors, "validation error"))
			} else {
				errors := []apputils.ErrorValidation{{
					Key:     "error",
					Message: "Invalid request body format: " + err.Error(),
				}}

				return c.Status(fiber.StatusBadRequest).JSON(apputils.ErrorValidationResponse(fiber.StatusBadRequest, errors, "invalid request format"))
			}
		}

		if err := validate.Struct(obj); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				errors := mappingErrorMessage(validationErrors)
				return c.Status(fiber.StatusBadRequest).JSON(apputils.ErrorValidationResponse(fiber.StatusBadRequest, errors, "validation error"))
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": fmt.Sprintf("validation error: %v", err),
				})
			}
		}

		c.Locals(constants.RequestBodyJSONKey, obj)
		return c.Next()
	}
}

// Mapping validation errors to more user-friendly messages
func mappingErrorMessage(validationErrors validator.ValidationErrors) []apputils.ErrorValidation {
	errors := make([]apputils.ErrorValidation, 0, len(validationErrors))

	for _, validationErr := range validationErrors {
		fieldName := apputils.StringToSnakeCase(validationErr.Field())
		tag := validationErr.Tag()
		param := validationErr.Param()

		// Create more helpful error messages based on validation tag
		var message string
		switch tag {
		case "required":
			message = fmt.Sprintf("Field %s wajib diisi", fieldName)
		case "required_without":
			message = fmt.Sprintf("Field %s wajib diisi ketika %s tidak tersedia", fieldName, param)
		case "required_with":
			message = fmt.Sprintf("Field %s wajib diisi ketika %s tersedia", fieldName, param)
		case "required_if":
			if param != "" {
				// param format is usually "field_name value", let's parse it
				parts := strings.Fields(param)
				if len(parts) >= 2 {
					dependentField := apputils.StringToSnakeCase(parts[0])
					dependentValue := strings.Join(parts[1:], " ")
					message = fmt.Sprintf("Field %s wajib diisi ketika field %s bernilai '%s'", fieldName, dependentField, dependentValue)
				} else {
					message = fmt.Sprintf("Field %s wajib diisi ketika kondisi %s terpenuhi", fieldName, param)
				}
			} else {
				message = fmt.Sprintf("Field %s wajib diisi ketika kondisi tertentu terpenuhi", fieldName)
			}
		case "gt":
			message = fmt.Sprintf("Field %s harus lebih besar dari %s", fieldName, param)
		case "gtfield":
			message = fmt.Sprintf("Field %s harus lebih besar dari field %s", fieldName, param)
		case "gte":
			message = fmt.Sprintf("Field %s harus lebih besar atau sama dengan %s", fieldName, param)
		case "gtefield":
			message = fmt.Sprintf("Field %s harus lebih besar atau sama dengan field %s", fieldName, param)
		case "lt":
			message = fmt.Sprintf("Field %s harus lebih kecil dari %s", fieldName, param)
		case "ltfield":
			message = fmt.Sprintf("Field %s harus lebih kecil dari field %s", fieldName, param)
		case "lte":
			message = fmt.Sprintf("Field %s harus lebih kecil atau sama dengan %s", fieldName, param)
		case "ltefield":
			message = fmt.Sprintf("Field %s harus lebih kecil atau sama dengan field %s", fieldName, param)
		case "eq":
			message = fmt.Sprintf("Field %s harus sama dengan %s", fieldName, param)
		case "eqfield":
			message = fmt.Sprintf("Field %s harus sama dengan field %s", fieldName, param)
		case "ne":
			message = fmt.Sprintf("Field %s tidak boleh sama dengan %s", fieldName, param)
		case "nefield":
			message = fmt.Sprintf("Field %s tidak boleh sama dengan field %s", fieldName, param)
		case "oneof":
			message = fmt.Sprintf("Field %s harus salah satu dari: %s", fieldName, param)
		case "email":
			message = fmt.Sprintf("Field %s harus berupa email yang valid", fieldName)
		case "url":
			message = fmt.Sprintf("Field %s harus berupa URL yang valid", fieldName)
		case "uri":
			message = fmt.Sprintf("Field %s harus berupa URI yang valid", fieldName)
		case "min":
			message = fmt.Sprintf("Field %s harus memiliki minimal %s karakter/item", fieldName, param)
		case "max":
			message = fmt.Sprintf("Field %s harus memiliki maksimal %s karakter/item", fieldName, param)
		case "len":
			message = fmt.Sprintf("Field %s harus memiliki panjang tepat %s karakter/item", fieldName, param)
		case "uuid":
			message = fmt.Sprintf("Field %s harus berupa UUID yang valid", fieldName)
		case "uuid3":
			message = fmt.Sprintf("Field %s harus berupa UUID versi 3 yang valid", fieldName)
		case "uuid4":
			message = fmt.Sprintf("Field %s harus berupa UUID versi 4 yang valid", fieldName)
		case "uuid5":
			message = fmt.Sprintf("Field %s harus berupa UUID versi 5 yang valid", fieldName)
		case "numeric":
			message = fmt.Sprintf("Field %s harus berupa angka", fieldName)
		case "number":
			message = fmt.Sprintf("Field %s harus berupa nomor yang valid", fieldName)
		case "alpha":
			message = fmt.Sprintf("Field %s hanya boleh berisi huruf", fieldName)
		case "alphanum":
			message = fmt.Sprintf("Field %s hanya boleh berisi huruf dan angka", fieldName)
		case "alphaunicode":
			message = fmt.Sprintf("Field %s hanya boleh berisi huruf unicode", fieldName)
		case "alphanumunicode":
			message = fmt.Sprintf("Field %s hanya boleh berisi huruf dan angka unicode", fieldName)
		case "boolean":
			message = fmt.Sprintf("Field %s harus berupa nilai boolean (true/false)", fieldName)
		case "json":
			message = fmt.Sprintf("Field %s harus berupa JSON yang valid", fieldName)
		case "base64":
			message = fmt.Sprintf("Field %s harus berupa base64 yang valid", fieldName)
		case "base64url":
			message = fmt.Sprintf("Field %s harus berupa base64url yang valid", fieldName)
		case "hexadecimal":
			message = fmt.Sprintf("Field %s harus berupa hexadecimal yang valid", fieldName)
		case "hexcolor":
			message = fmt.Sprintf("Field %s harus berupa kode warna hex yang valid", fieldName)
		case "rgb":
			message = fmt.Sprintf("Field %s harus berupa kode warna RGB yang valid", fieldName)
		case "rgba":
			message = fmt.Sprintf("Field %s harus berupa kode warna RGBA yang valid", fieldName)
		case "hsl":
			message = fmt.Sprintf("Field %s harus berupa kode warna HSL yang valid", fieldName)
		case "hsla":
			message = fmt.Sprintf("Field %s harus berupa kode warna HSLA yang valid", fieldName)
		case "ip":
			message = fmt.Sprintf("Field %s harus berupa alamat IP yang valid", fieldName)
		case "ipv4":
			message = fmt.Sprintf("Field %s harus berupa alamat IPv4 yang valid", fieldName)
		case "ipv6":
			message = fmt.Sprintf("Field %s harus berupa alamat IPv6 yang valid", fieldName)
		case "mac":
			message = fmt.Sprintf("Field %s harus berupa alamat MAC yang valid", fieldName)
		case "latitude":
			message = fmt.Sprintf("Field %s harus berupa latitude yang valid", fieldName)
		case "longitude":
			message = fmt.Sprintf("Field %s harus berupa longitude yang valid", fieldName)
		case "datetime":
			message = fmt.Sprintf("Field %s harus berupa format tanggal dan waktu yang valid", fieldName)
		case "timezone":
			message = fmt.Sprintf("Field %s harus berupa timezone yang valid", fieldName)
		case "isbn":
			message = fmt.Sprintf("Field %s harus berupa ISBN yang valid", fieldName)
		case "isbn10":
			message = fmt.Sprintf("Field %s harus berupa ISBN-10 yang valid", fieldName)
		case "isbn13":
			message = fmt.Sprintf("Field %s harus berupa ISBN-13 yang valid", fieldName)
		case "credit_card":
			message = fmt.Sprintf("Field %s harus berupa nomor kartu kredit yang valid", fieldName)
		case "ssn":
			message = fmt.Sprintf("Field %s harus berupa nomor SSN yang valid", fieldName)
		case "contains":
			message = fmt.Sprintf("Field %s harus mengandung '%s'", fieldName, param)
		case "containsany":
			message = fmt.Sprintf("Field %s harus mengandung salah satu dari karakter: %s", fieldName, param)
		case "containsrune":
			message = fmt.Sprintf("Field %s harus mengandung karakter '%s'", fieldName, param)
		case "excludes":
			message = fmt.Sprintf("Field %s tidak boleh mengandung '%s'", fieldName, param)
		case "excludesall":
			message = fmt.Sprintf("Field %s tidak boleh mengandung karakter: %s", fieldName, param)
		case "excludesrune":
			message = fmt.Sprintf("Field %s tidak boleh mengandung karakter '%s'", fieldName, param)
		case "startswith":
			message = fmt.Sprintf("Field %s harus dimulai dengan '%s'", fieldName, param)
		case "endswith":
			message = fmt.Sprintf("Field %s harus diakhiri dengan '%s'", fieldName, param)
		case "lowercase":
			message = fmt.Sprintf("Field %s harus berupa huruf kecil semua", fieldName)
		case "uppercase":
			message = fmt.Sprintf("Field %s harus berupa huruf besar semua", fieldName)
		case "file":
			message = fmt.Sprintf("Field %s harus berupa file yang valid", fieldName)
		case "dir":
			message = fmt.Sprintf("Field %s harus berupa direktori yang valid", fieldName)
		case "unique":
			message = fmt.Sprintf("Field %s harus memiliki nilai yang unik", fieldName)
		case "ascii":
			message = fmt.Sprintf("Field %s hanya boleh berisi karakter ASCII", fieldName)
		case "printascii":
			message = fmt.Sprintf("Field %s hanya boleh berisi karakter ASCII yang dapat dicetak", fieldName)
		case "multibyte":
			message = fmt.Sprintf("Field %s harus berisi karakter multibyte", fieldName)
		case "datauri":
			message = fmt.Sprintf("Field %s harus berupa data URI yang valid", fieldName)
		case "hostname":
			message = fmt.Sprintf("Field %s harus berupa hostname yang valid", fieldName)
		case "hostname_rfc1123":
			message = fmt.Sprintf("Field %s harus berupa hostname RFC1123 yang valid", fieldName)
		case "fqdn":
			message = fmt.Sprintf("Field %s harus berupa FQDN yang valid", fieldName)
		case "html":
			message = fmt.Sprintf("Field %s harus berupa HTML yang valid", fieldName)
		case "html_encoded":
			message = fmt.Sprintf("Field %s harus berupa HTML yang ter-encode", fieldName)
		case "url_encoded":
			message = fmt.Sprintf("Field %s harus berupa URL yang ter-encode", fieldName)
		case "dive":
			message = fmt.Sprintf("Field %s memiliki elemen yang tidak valid", fieldName)
		default:
			message = fmt.Sprintf("Field %s tidak valid: %s", fieldName, validationErr.Error())
		}

		errors = append(errors, apputils.ErrorValidation{
			Key:     fieldName,
			Message: message,
		})
	}

	return errors
}
