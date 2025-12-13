package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTinyDBParser_ParseData(t *testing.T) {
	parser := NewTinyDBParser()

	t.Run("valid document", func(t *testing.T) {
		jsonData := []byte(`{
			"activities": {
				"1": {
					"_id": 1,
					"name": "Test Event",
					"description": "A test event",
					"date": "2024-06-15T10:00:00",
					"location": "Test Location",
					"activity_type": "event",
					"status": "active",
					"requires_registration": true,
					"is_free": false,
					"fee": 25.50,
					"max_capacity": 50,
					"created_at": "2024-01-01T00:00:00",
					"updated_at": "2024-01-01T00:00:00"
				}
			},
			"people": {
				"1": {
					"_id": 1,
					"first_name": "John",
					"last_name": "Doe",
					"email": "john.doe@example.com",
					"phone": "555-1234",
					"is_active": true,
					"created_at": "2024-01-01T00:00:00",
					"updated_at": "2024-01-01T00:00:00"
				}
			},
			"attendees": {
				"1": {
					"_id": 1,
					"activity_id": 1,
					"person_id": 1,
					"role": "participant",
					"payment_status": "paid",
					"payment_amount": 25.50,
					"payment_date": "2024-01-05T00:00:00",
					"registration_date": "2024-01-01T00:00:00"
				}
			}
		}`)

		doc, err := parser.ParseData(jsonData)
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Len(t, doc.Activities, 1)
		assert.Len(t, doc.People, 1)
		assert.Len(t, doc.Attendees, 1)
	})

	t.Run("empty document", func(t *testing.T) {
		jsonData := []byte(`{
			"activities": {},
			"people": {},
			"attendees": {}
		}`)

		doc, err := parser.ParseData(jsonData)
		require.NoError(t, err)
		assert.NotNil(t, doc)
		assert.Len(t, doc.Activities, 0)
		assert.Len(t, doc.People, 0)
		assert.Len(t, doc.Attendees, 0)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		jsonData := []byte(`{invalid json}`)

		_, err := parser.ParseData(jsonData)
		assert.Error(t, err)
	})
}

func TestTinyDBParser_Validate(t *testing.T) {
	parser := NewTinyDBParser()

	t.Run("valid document", func(t *testing.T) {
		doc := &TinyDBDocument{
			Activities: map[string]TinyDBActivity{
				"1": {
					DocID:    1,
					Name:     "Test Event",
					Location: "Test Location",
					Date:     "2024-06-15T10:00:00",
				},
			},
			People: map[string]TinyDBPerson{
				"1": {
					DocID:     1,
					FirstName: "John",
					LastName:  "Doe",
				},
			},
			Attendees: map[string]TinyDBAttendee{
				"1": {
					DocID:      1,
					ActivityID: 1,
					PersonID:   1,
				},
			},
		}

		err := parser.Validate(doc)
		assert.NoError(t, err)
	})

	t.Run("activity missing name", func(t *testing.T) {
		doc := &TinyDBDocument{
			Activities: map[string]TinyDBActivity{
				"1": {
					DocID:    1,
					Name:     "",
					Location: "Test Location",
					Date:     "2024-06-15",
				},
			},
			People:    map[string]TinyDBPerson{},
			Attendees: map[string]TinyDBAttendee{},
		}

		err := parser.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty name")
	})

	t.Run("activity invalid date", func(t *testing.T) {
		doc := &TinyDBDocument{
			Activities: map[string]TinyDBActivity{
				"1": {
					DocID:    1,
					Name:     "Test",
					Location: "Test Location",
					Date:     "invalid-date",
				},
			},
			People:    map[string]TinyDBPerson{},
			Attendees: map[string]TinyDBAttendee{},
		}

		err := parser.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
	})

	t.Run("person missing first name", func(t *testing.T) {
		doc := &TinyDBDocument{
			Activities: map[string]TinyDBActivity{},
			People: map[string]TinyDBPerson{
				"1": {
					DocID:     1,
					FirstName: "",
					LastName:  "Doe",
				},
			},
			Attendees: map[string]TinyDBAttendee{},
		}

		err := parser.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty first name")
	})

	t.Run("attendee zero activity ID", func(t *testing.T) {
		doc := &TinyDBDocument{
			Activities: map[string]TinyDBActivity{},
			People:     map[string]TinyDBPerson{},
			Attendees: map[string]TinyDBAttendee{
				"1": {
					DocID:      1,
					ActivityID: 0,
					PersonID:   1,
				},
			},
		}

		err := parser.Validate(doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero activity ID")
	})
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"RFC3339", "2024-01-15T10:30:00Z", false},
		{"RFC3339 no Z", "2024-01-15T10:30:00", false},
		{"Date with space", "2024-01-15 10:30:00", false},
		{"Date only", "2024-01-15", false},
		{"RFC3339Nano", "2024-01-15T10:30:00.123456789Z", false},
		{"Invalid", "not-a-date", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDateTime(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
