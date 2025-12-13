package domain

import (
	"strings"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// Person represents a contact in the database
type Person struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPerson creates a new Person with the given details
func NewPerson(firstName, lastName, email, phone string) *Person {
	return &Person{
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
		Email:     util.NormalizeEmail(email),
		Phone:     util.NormalizePhone(phone),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// FullName returns the person's full name
func (p *Person) FullName() string {
	return strings.TrimSpace(p.FirstName + " " + p.LastName)
}

// DisplayName returns a display-friendly name
func (p *Person) DisplayName() string {
	return p.FullName()
}

// HasEmail returns true if the person has an email address
func (p *Person) HasEmail() bool {
	return strings.TrimSpace(p.Email) != ""
}

// HasPhone returns true if the person has a phone number
func (p *Person) HasPhone() bool {
	return strings.TrimSpace(p.Phone) != ""
}

// HasContactInfo returns true if the person has at least one contact method
func (p *Person) HasContactInfo() bool {
	return p.HasEmail() || p.HasPhone()
}

// GetPrimaryContact returns the primary contact method (email preferred)
func (p *Person) GetPrimaryContact() string {
	if p.HasEmail() {
		return p.Email
	}
	if p.HasPhone() {
		return p.Phone
	}
	return ""
}

// UpdateEmail updates the person's email address
func (p *Person) UpdateEmail(email string) error {
	normalizedEmail := util.NormalizeEmail(email)

	// Validate email format
	if normalizedEmail != "" {
		if err := util.ValidateEmail(normalizedEmail); err != nil {
			return err
		}
	}

	// Ensure at least one contact method remains
	if normalizedEmail == "" && !p.HasPhone() {
		return util.ErrPersonContactRequired
	}

	p.Email = normalizedEmail
	p.UpdatedAt = time.Now()
	return nil
}

// UpdatePhone updates the person's phone number
func (p *Person) UpdatePhone(phone string) error {
	normalizedPhone := util.NormalizePhone(phone)

	// Validate phone format
	if normalizedPhone != "" {
		if err := util.ValidatePhone(phone); err != nil {
			return err
		}
	}

	// Ensure at least one contact method remains
	if normalizedPhone == "" && !p.HasEmail() {
		return util.ErrPersonContactRequired
	}

	p.Phone = normalizedPhone
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateName updates the person's name
func (p *Person) UpdateName(firstName, lastName string) error {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if firstName == "" {
		return util.ErrPersonFirstNameRequired
	}

	if lastName == "" {
		return util.ErrPersonLastNameRequired
	}

	p.FirstName = firstName
	p.LastName = lastName
	p.UpdatedAt = time.Now()
	return nil
}

// Normalize normalizes all person data (trim, lowercase email, etc.)
func (p *Person) Normalize() {
	p.FirstName = strings.TrimSpace(p.FirstName)
	p.LastName = strings.TrimSpace(p.LastName)
	p.Email = util.NormalizeEmail(p.Email)
	p.Phone = util.NormalizePhone(p.Phone)
}

// Validate validates the person data
func (p *Person) Validate() error {
	// Normalize data first
	p.Normalize()

	// Required fields
	if err := util.ValidateRequired(p.FirstName, "first_name"); err != nil {
		return err
	}

	if err := util.ValidateRequired(p.LastName, "last_name"); err != nil {
		return err
	}

	// String length constraints
	if err := util.ValidateStringLength(p.FirstName, "first_name", 1, 100); err != nil {
		return err
	}

	if err := util.ValidateStringLength(p.LastName, "last_name", 1, 100); err != nil {
		return err
	}

	// At least one contact method is required
	if !p.HasContactInfo() {
		return util.ErrPersonContactRequired
	}

	// Validate email format if provided
	if p.HasEmail() {
		if err := util.ValidateEmail(p.Email); err != nil {
			return err
		}

		// Email length constraint
		if err := util.ValidateStringLength(p.Email, "email", 0, 255); err != nil {
			return err
		}
	}

	// Validate phone format if provided
	if p.HasPhone() {
		if err := util.ValidatePhone(p.Phone); err != nil {
			return err
		}

		// Phone length constraint (after normalization)
		if err := util.ValidateStringLength(p.Phone, "phone", 0, 20); err != nil {
			return err
		}
	}

	return nil
}

// Equals checks if two persons are equal (by ID or matching details)
func (p *Person) Equals(other *Person) bool {
	if other == nil {
		return false
	}

	// If both have IDs, compare by ID
	if p.ID != 0 && other.ID != 0 {
		return p.ID == other.ID
	}

	// Otherwise, compare by details
	return p.FirstName == other.FirstName &&
		p.LastName == other.LastName &&
		p.Email == other.Email &&
		p.Phone == other.Phone
}

// IsDuplicateOf checks if this person is a potential duplicate of another
// (same name and at least one matching contact method)
func (p *Person) IsDuplicateOf(other *Person) bool {
	if other == nil {
		return false
	}

	// Same name
	sameName := strings.EqualFold(p.FirstName, other.FirstName) &&
		strings.EqualFold(p.LastName, other.LastName)

	if !sameName {
		return false
	}

	// At least one matching contact method
	sameEmail := p.HasEmail() && other.HasEmail() &&
		strings.EqualFold(p.Email, other.Email)

	samePhone := p.HasPhone() && other.HasPhone() &&
		p.Phone == other.Phone

	return sameEmail || samePhone
}
