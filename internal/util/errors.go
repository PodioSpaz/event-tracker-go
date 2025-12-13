package util

import "errors"

// Domain model validation errors
var (
	// Activity errors
	ErrActivityNameRequired     = errors.New("activity name is required")
	ErrActivityLocationRequired = errors.New("activity location is required")
	ErrActivityDateRequired     = errors.New("activity date is required")
	ErrActivityDateInvalid      = errors.New("activity date is invalid")
	ErrActivityFeeNegative      = errors.New("activity fee cannot be negative")
	ErrActivityTypeInvalid      = errors.New("activity type is invalid")
	ErrActivityStatusInvalid    = errors.New("activity status is invalid")
	ErrActivityCapacityNegative = errors.New("activity capacity cannot be negative")
	ErrActivityNotFound         = errors.New("activity not found")

	// Person errors
	ErrPersonFirstNameRequired = errors.New("person first name is required")
	ErrPersonLastNameRequired  = errors.New("person last name is required")
	ErrPersonContactRequired   = errors.New("person must have at least one contact method (email or phone)")
	ErrPersonInvalidEmail      = errors.New("person email is invalid")
	ErrPersonInvalidPhone      = errors.New("person phone number is invalid")
	ErrPersonNotFound          = errors.New("person not found")
	ErrPersonDuplicate         = errors.New("person with this email already exists")

	// Attendee errors
	ErrAttendeeActivityRequired  = errors.New("attendee must be associated with an activity")
	ErrAttendeePersonRequired    = errors.New("attendee must be associated with a person")
	ErrAttendeeRoleInvalid       = errors.New("attendee role is invalid")
	ErrAttendeePaymentNegative   = errors.New("attendee payment amount cannot be negative")
	ErrAttendeePaymentStatusInvalid = errors.New("attendee payment status is invalid")
	ErrAttendeeNotFound          = errors.New("attendee not found")
	ErrAttendeeAlreadyRegistered = errors.New("person is already registered for this activity")
	ErrAttendeeHasRegistrations  = errors.New("person has existing registrations and cannot be deleted")

	// Business rule errors
	ErrActivityFull              = errors.New("activity is at full capacity")
	ErrActivityNotAcceptingRegistrations = errors.New("activity is not accepting registrations")
	ErrActivityCancelled         = errors.New("activity is cancelled")
	ErrActivityCompleted         = errors.New("activity is already completed")
	ErrPaymentAmountMismatch     = errors.New("payment amount does not match event fee")
	ErrCannotRegisterForGathering = errors.New("cannot register attendees for gatherings")

	// Validation errors
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidPhone       = errors.New("invalid phone number format")
	ErrInvalidDate        = errors.New("invalid date format")
	ErrInvalidDecimal     = errors.New("invalid decimal value")
	ErrRequiredField      = errors.New("required field is missing")
	ErrValueTooLong       = errors.New("value exceeds maximum length")
	ErrValueTooShort      = errors.New("value is below minimum length")
	ErrValueOutOfRange    = errors.New("value is out of valid range")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseTransaction = errors.New("database transaction failed")
	ErrRecordNotFound     = errors.New("record not found")
	ErrDuplicateRecord    = errors.New("duplicate record")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")

	// General errors
	ErrInternal       = errors.New("internal error")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrInvalidInput   = errors.New("invalid input")
	ErrOperationFailed = errors.New("operation failed")
)

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// BusinessRuleError represents a business rule violation
type BusinessRuleError struct {
	Rule    string
	Message string
	Err     error
}

func (e *BusinessRuleError) Error() string {
	if e.Rule != "" {
		return e.Rule + ": " + e.Message
	}
	return e.Message
}

func (e *BusinessRuleError) Unwrap() error {
	return e.Err
}

// NewBusinessRuleError creates a new business rule error
func NewBusinessRuleError(rule, message string, err error) *BusinessRuleError {
	return &BusinessRuleError{
		Rule:    rule,
		Message: message,
		Err:     err,
	}
}
