package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPerson(t *testing.T) {
	person := NewPerson("John", "Doe", "john.doe@example.com", "555-1234")

	assert.NotNil(t, person)
	assert.Equal(t, "John", person.FirstName)
	assert.Equal(t, "Doe", person.LastName)
	assert.Equal(t, "john.doe@example.com", person.Email)
	assert.NotEmpty(t, person.Phone)
}

func TestPerson_FullName(t *testing.T) {
	person := &Person{
		FirstName: "John",
		LastName:  "Doe",
	}

	assert.Equal(t, "John Doe", person.FullName())
}

func TestPerson_HasEmail(t *testing.T) {
	person := &Person{Email: "test@example.com"}
	assert.True(t, person.HasEmail())

	person.Email = ""
	assert.False(t, person.HasEmail())

	person.Email = "  "
	assert.False(t, person.HasEmail())
}

func TestPerson_HasPhone(t *testing.T) {
	person := &Person{Phone: "5551234567"}
	assert.True(t, person.HasPhone())

	person.Phone = ""
	assert.False(t, person.HasPhone())
}

func TestPerson_HasContactInfo(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		phone    string
		expected bool
	}{
		{"Has email only", "test@example.com", "", true},
		{"Has phone only", "", "5551234567", true},
		{"Has both", "test@example.com", "5551234567", true},
		{"Has neither", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			person := &Person{Email: tt.email, Phone: tt.phone}
			assert.Equal(t, tt.expected, person.HasContactInfo())
		})
	}
}

func TestPerson_GetPrimaryContact(t *testing.T) {
	// Email preferred
	person := &Person{Email: "test@example.com", Phone: "5551234567"}
	assert.Equal(t, "test@example.com", person.GetPrimaryContact())

	// Phone if no email
	person = &Person{Phone: "5551234567"}
	assert.Equal(t, "5551234567", person.GetPrimaryContact())

	// Empty if neither
	person = &Person{}
	assert.Equal(t, "", person.GetPrimaryContact())
}

func TestPerson_UpdateEmail(t *testing.T) {
	person := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "old@example.com",
		Phone:     "5551234567",
	}

	// Valid update
	err := person.UpdateEmail("new@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "new@example.com", person.Email)

	// Invalid format
	err = person.UpdateEmail("invalid-email")
	assert.Error(t, err)

	// Cannot remove email if no phone
	person.Phone = ""
	err = person.UpdateEmail("")
	assert.Error(t, err)

	// Can remove email if has phone
	person.Phone = "5551234567"
	err = person.UpdateEmail("")
	assert.NoError(t, err)
	assert.Equal(t, "", person.Email)
}

func TestPerson_UpdatePhone(t *testing.T) {
	person := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "test@example.com",
		Phone:     "5551234567",
	}

	// Valid update
	err := person.UpdatePhone("555-999-8888")
	assert.NoError(t, err)
	assert.NotEmpty(t, person.Phone)

	// Cannot remove phone if no email
	person.Email = ""
	err = person.UpdatePhone("")
	assert.Error(t, err)

	// Can remove phone if has email
	person.Email = "test@example.com"
	err = person.UpdatePhone("")
	assert.NoError(t, err)
	assert.Equal(t, "", person.Phone)
}

func TestPerson_UpdateName(t *testing.T) {
	person := &Person{FirstName: "John", LastName: "Doe"}

	// Valid update
	err := person.UpdateName("Jane", "Smith")
	assert.NoError(t, err)
	assert.Equal(t, "Jane", person.FirstName)
	assert.Equal(t, "Smith", person.LastName)

	// Empty first name
	err = person.UpdateName("", "Smith")
	assert.Error(t, err)

	// Empty last name
	err = person.UpdateName("Jane", "")
	assert.Error(t, err)
}

func TestPerson_Normalize(t *testing.T) {
	person := &Person{
		FirstName: "  John  ",
		LastName:  "  Doe  ",
		Email:     " TEST@EXAMPLE.COM ",
		Phone:     " 555-123-4567 ",
	}

	person.Normalize()

	assert.Equal(t, "John", person.FirstName)
	assert.Equal(t, "Doe", person.LastName)
	assert.Equal(t, "test@example.com", person.Email)
	assert.NotContains(t, person.Phone, " ")
	assert.NotContains(t, person.Phone, "-")
}

func TestPerson_Validate(t *testing.T) {
	validPerson := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	err := validPerson.Validate()
	assert.NoError(t, err)

	// Missing first name
	invalidPerson := *validPerson
	invalidPerson.FirstName = ""
	assert.Error(t, invalidPerson.Validate())

	// Missing last name
	invalidPerson = *validPerson
	invalidPerson.LastName = ""
	assert.Error(t, invalidPerson.Validate())

	// No contact info
	invalidPerson = *validPerson
	invalidPerson.Email = ""
	invalidPerson.Phone = ""
	assert.Error(t, invalidPerson.Validate())

	// Invalid email
	invalidPerson = *validPerson
	invalidPerson.Email = "not-an-email"
	assert.Error(t, invalidPerson.Validate())

	// Valid with phone only
	validWithPhone := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "5551234567",
	}
	assert.NoError(t, validWithPhone.Validate())
}

func TestPerson_Equals(t *testing.T) {
	person1 := &Person{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	person2 := &Person{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	person3 := &Person{
		ID:        2,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	// Same ID
	assert.True(t, person1.Equals(person2))

	// Different ID
	assert.False(t, person1.Equals(person3))

	// Same details, no ID
	person1.ID = 0
	person2.ID = 0
	assert.True(t, person1.Equals(person2))

	// Nil comparison
	assert.False(t, person1.Equals(nil))
}

func TestPerson_IsDuplicateOf(t *testing.T) {
	person1 := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "5551234567",
	}

	// Same name and email
	person2 := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "9999999999",
	}
	assert.True(t, person1.IsDuplicateOf(person2))

	// Same name and phone
	person3 := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "different@example.com",
		Phone:     "5551234567",
	}
	assert.True(t, person1.IsDuplicateOf(person3))

	// Same name, different contact
	person4 := &Person{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "different@example.com",
		Phone:     "9999999999",
	}
	assert.False(t, person1.IsDuplicateOf(person4))

	// Different name
	person5 := &Person{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "john@example.com",
		Phone:     "5551234567",
	}
	assert.False(t, person1.IsDuplicateOf(person5))

	// Nil comparison
	assert.False(t, person1.IsDuplicateOf(nil))
}
