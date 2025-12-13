package util

import (
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var (
	// Email validation regex - RFC 5322 simplified
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone validation regex - allows various formats
	// Matches: +1234567890, 123-456-7890, (123) 456-7890, etc.
	phoneRegex = regexp.MustCompile(`^\+?\d{7,15}$`)

	// Formatting characters to strip from phone numbers
	phoneFormatChars = regexp.MustCompile(`[\s\-\(\)\+\.]`)
)

// ValidateEmail validates an email address format
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Empty is allowed, caller decides if required
	}

	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePhone validates a phone number format
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Empty is allowed, caller decides if required
	}

	// Strip formatting characters
	cleaned := StripPhoneFormatting(phone)

	if !phoneRegex.MatchString(cleaned) {
		return ErrInvalidPhone
	}

	return nil
}

// StripPhoneFormatting removes common phone formatting characters
func StripPhoneFormatting(phone string) string {
	return phoneFormatChars.ReplaceAllString(phone, "")
}

// ValidateRequired checks if a string value is non-empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(fieldName, "is required", ErrRequiredField)
	}
	return nil
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(strings.TrimSpace(value))

	if min > 0 && length < min {
		return NewValidationError(fieldName, "is too short", ErrValueTooShort)
	}

	if max > 0 && length > max {
		return NewValidationError(fieldName, "is too long", ErrValueTooLong)
	}

	return nil
}

// ValidateDate validates a date is not zero
func ValidateDate(date time.Time, fieldName string) error {
	if date.IsZero() {
		return NewValidationError(fieldName, "is required", ErrInvalidDate)
	}
	return nil
}

// ValidateDateInFuture validates a date is in the future
func ValidateDateInFuture(date time.Time, fieldName string) error {
	if err := ValidateDate(date, fieldName); err != nil {
		return err
	}

	if date.Before(time.Now()) {
		return NewValidationError(fieldName, "must be in the future", ErrInvalidDate)
	}

	return nil
}

// ValidateDateInPast validates a date is in the past
func ValidateDateInPast(date time.Time, fieldName string) error {
	if err := ValidateDate(date, fieldName); err != nil {
		return err
	}

	if date.After(time.Now()) {
		return NewValidationError(fieldName, "must be in the past", ErrInvalidDate)
	}

	return nil
}

// ValidateDecimal validates a decimal value
func ValidateDecimal(value decimal.Decimal, fieldName string, allowNegative bool) error {
	if !allowNegative && value.IsNegative() {
		return NewValidationError(fieldName, "cannot be negative", ErrInvalidDecimal)
	}
	return nil
}

// ValidateDecimalInRange validates a decimal is within a range
func ValidateDecimalInRange(value decimal.Decimal, fieldName string, min, max decimal.Decimal) error {
	if value.LessThan(min) || value.GreaterThan(max) {
		return NewValidationError(fieldName, "is out of valid range", ErrValueOutOfRange)
	}
	return nil
}

// ValidateIntInRange validates an integer is within a range
func ValidateIntInRange(value int, fieldName string, min, max int) error {
	if value < min || value > max {
		return NewValidationError(fieldName, "is out of valid range", ErrValueOutOfRange)
	}
	return nil
}

// ValidatePositiveInt validates an integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
	if value < 0 {
		return NewValidationError(fieldName, "must be positive", ErrValueOutOfRange)
	}
	return nil
}

// NormalizeEmail normalizes an email address (lowercase, trim)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizePhone normalizes a phone number (strips formatting)
func NormalizePhone(phone string) string {
	return StripPhoneFormatting(strings.TrimSpace(phone))
}

// ParseDecimal parses a string to decimal.Decimal
func ParseDecimal(value string) (decimal.Decimal, error) {
	if value == "" {
		return decimal.Zero, nil
	}

	d, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, ErrInvalidDecimal
	}

	return d, nil
}

// FormatDecimal formats a decimal for storage (2 decimal places)
func FormatDecimal(value decimal.Decimal) string {
	return value.StringFixed(2)
}

// DecimalFromFloat creates a decimal from float64 with 2 decimal places
func DecimalFromFloat(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value).Round(2)
}
