// internal/security/validatior.go

package security

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationError struct{ msg string }

func (e *ValidationError) Error() string { return e.msg }

func newValidationErr(format string, a ...any) error {
	return &ValidationError{msg: fmt.Sprintf(format, a...)}
}

var (
	nameRe = regexp.MustCompile(`^[A-Za-z.\s]+$`)
	rollRe = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
	ctrlRe = regexp.MustCompile(`[\x00-\x1f\x7f]`)
)

func SanitizeText(value string) string {
	value = strings.TrimSpace(value)
	return ctrlRe.ReplaceAllString(value, "")
}

func ValidateName(name string, maxLen int) (string, error) {
	name = SanitizeText(name)
	if len(name) < 2 || len(name) > maxLen {
		return "", newValidationErr("name length must be between 2 and %d characters", maxLen)
	}
	if !nameRe.MatchString(name) {
		return "", newValidationErr("name contains invalid characters")
	}
	return name, nil
}

func ValidateRoll(roll string, maxLen int) (string, error) {
	roll = SanitizeText(roll)
	if len(roll) < 1 || len(roll) > maxLen {
		return "", newValidationErr("roll number length must be between 1 and %d characters", maxLen)
	}
	if !rollRe.MatchString(roll) {
		return "", newValidationErr("roll number must be alphanumeric only")
	}
	return strings.ToUpper(roll), nil
}

func ValidateSessionCode(code, expected string) error {
	code = SanitizeText(code)
	if code != expected {
		return newValidationErr("invalid session code")
	}
	return nil
}
