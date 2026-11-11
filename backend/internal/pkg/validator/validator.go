package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	// Register a custom validation rule: "valid-sku"
	// Usage in struct tags: validate:"required,valid-sku"
	_ = Validate.RegisterValidation("valid-sku", validateSKU)
}

// Custom validation function checking if string starts with "SKU-"
func validateSKU(fl validator.FieldLevel) bool {
	sku, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return strings.HasPrefix(sku, "SKU-")
}

func FormatErrors(err error) map[string]string {
	errorsMap := make(map[string]string)
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			switch fieldErr.Tag() {
			case "valid-sku":
				errorsMap[fieldErr.Field()] = "must start with 'SKU-'"
			default:
				errorsMap[fieldErr.Field()] = "failed on validation rule: " + fieldErr.Tag()
			}
		}
	} else {
		errorsMap["general"] = err.Error()
	}
	return errorsMap
}
