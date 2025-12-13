package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TinyDBDocument represents a TinyDB JSON document structure
type TinyDBDocument struct {
	Activities map[string]TinyDBActivity `json:"activities"`
	People     map[string]TinyDBPerson   `json:"people"`
	Attendees  map[string]TinyDBAttendee `json:"attendees"`
}

// TinyDBActivity represents an activity in TinyDB format
type TinyDBActivity struct {
	DocID                int     `json:"_id"`
	Name                 string  `json:"name"`
	Description          string  `json:"description,omitempty"`
	Date                 string  `json:"date"`
	Location             string  `json:"location"`
	ActivityType         string  `json:"activity_type"`
	Status               string  `json:"status"`
	RequiresRegistration bool    `json:"requires_registration"`
	IsFree               bool    `json:"is_free"`
	Fee                  float64 `json:"fee,omitempty"`
	MaxCapacity          *int    `json:"max_capacity,omitempty"`
	EstimatedHeadCount   *int    `json:"estimated_head_count,omitempty"`
	ActualHeadCount      *int    `json:"actual_head_count,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// TinyDBPerson represents a person in TinyDB format
type TinyDBPerson struct {
	DocID     int    `json:"_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Notes     string `json:"notes,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TinyDBAttendee represents an attendee in TinyDB format
type TinyDBAttendee struct {
	DocID            int     `json:"_id"`
	ActivityID       int     `json:"activity_id"`
	PersonID         int     `json:"person_id"`
	Role             string  `json:"role"`
	PaymentStatus    string  `json:"payment_status"`
	PaymentAmount    float64 `json:"payment_amount,omitempty"`
	PaymentDate      string  `json:"payment_date,omitempty"`
	RegistrationDate string  `json:"registration_date"`
	Notes            string  `json:"notes,omitempty"`
}

// TinyDBParser handles parsing of TinyDB JSON files
type TinyDBParser struct{}

// NewTinyDBParser creates a new TinyDB parser
func NewTinyDBParser() *TinyDBParser {
	return &TinyDBParser{}
}

// ParseFile parses a TinyDB JSON file
func (p *TinyDBParser) ParseFile(filePath string) (*TinyDBDocument, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var doc TinyDBDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &doc, nil
}

// ParseData parses TinyDB JSON data from bytes
func (p *TinyDBParser) ParseData(data []byte) (*TinyDBDocument, error) {
	var doc TinyDBDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &doc, nil
}

// Validate validates the TinyDB document structure
func (p *TinyDBParser) Validate(doc *TinyDBDocument) error {
	if doc == nil {
		return fmt.Errorf("document is nil")
	}

	// Validate activities
	for id, activity := range doc.Activities {
		if activity.Name == "" {
			return fmt.Errorf("activity %s has empty name", id)
		}
		if activity.Location == "" {
			return fmt.Errorf("activity %s has empty location", id)
		}
		if activity.Date == "" {
			return fmt.Errorf("activity %s has empty date", id)
		}
		// Validate date format
		if _, err := parseDateTime(activity.Date); err != nil {
			return fmt.Errorf("activity %s has invalid date format: %w", id, err)
		}
	}

	// Validate people
	for id, person := range doc.People {
		if person.FirstName == "" {
			return fmt.Errorf("person %s has empty first name", id)
		}
		if person.LastName == "" {
			return fmt.Errorf("person %s has empty last name", id)
		}
	}

	// Validate attendees
	for id, attendee := range doc.Attendees {
		if attendee.ActivityID == 0 {
			return fmt.Errorf("attendee %s has zero activity ID", id)
		}
		if attendee.PersonID == 0 {
			return fmt.Errorf("attendee %s has zero person ID", id)
		}
	}

	return nil
}

// parseDateTime parses various datetime formats from Python
func parseDateTime(dateStr string) (time.Time, error) {
	// Try common formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
