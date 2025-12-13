package migration

import (
	"fmt"
	"strings"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/shopspring/decimal"
)

// Mapper converts TinyDB structures to domain models
type Mapper struct{}

// NewMapper creates a new mapper
func NewMapper() *Mapper {
	return &Mapper{}
}

// MapActivity converts a TinyDB activity to a domain Activity
func (m *Mapper) MapActivity(tdb *TinyDBActivity) (*domain.Activity, error) {
	if tdb == nil {
		return nil, fmt.Errorf("tinydb activity is nil")
	}

	// Parse date
	date, err := parseDateTime(tdb.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse activity date: %w", err)
	}

	// Create activity
	activity := &domain.Activity{
		Name:                 tdb.Name,
		Description:          tdb.Description,
		Date:                 date,
		Location:             tdb.Location,
		RequiresRegistration: tdb.RequiresRegistration,
		IsFree:               tdb.IsFree,
		Fee:                  decimal.NewFromFloat(tdb.Fee),
		MaxCapacity:          tdb.MaxCapacity,
		EstimatedHeadCount:   tdb.EstimatedHeadCount,
		ActualHeadCount:      tdb.ActualHeadCount,
	}

	// Map activity type
	activityType, err := m.mapActivityType(tdb.ActivityType)
	if err != nil {
		return nil, err
	}
	activity.ActivityType = activityType

	// Map status
	status, err := m.mapActivityStatus(tdb.Status)
	if err != nil {
		return nil, err
	}
	activity.Status = status

	// Parse timestamps
	if tdb.CreatedAt != "" {
		createdAt, err := parseDateTime(tdb.CreatedAt)
		if err == nil {
			activity.CreatedAt = createdAt
		}
	}

	if tdb.UpdatedAt != "" {
		updatedAt, err := parseDateTime(tdb.UpdatedAt)
		if err == nil {
			activity.UpdatedAt = updatedAt
		}
	}

	// Set defaults if timestamps are zero
	if activity.CreatedAt.IsZero() {
		activity.CreatedAt = time.Now()
	}
	if activity.UpdatedAt.IsZero() {
		activity.UpdatedAt = time.Now()
	}

	return activity, nil
}

// MapPerson converts a TinyDB person to a domain Person
func (m *Mapper) MapPerson(tdb *TinyDBPerson) (*domain.Person, error) {
	if tdb == nil {
		return nil, fmt.Errorf("tinydb person is nil")
	}

	// Create person
	person := domain.NewPerson(tdb.FirstName, tdb.LastName, tdb.Email, tdb.Phone)
	// Note: Notes and IsActive fields are not in the current Person model
	// They may be added in a future version

	// Parse timestamps
	if tdb.CreatedAt != "" {
		createdAt, err := parseDateTime(tdb.CreatedAt)
		if err == nil {
			person.CreatedAt = createdAt
		}
	}

	if tdb.UpdatedAt != "" {
		updatedAt, err := parseDateTime(tdb.UpdatedAt)
		if err == nil {
			person.UpdatedAt = updatedAt
		}
	}

	// Set defaults if timestamps are zero
	if person.CreatedAt.IsZero() {
		person.CreatedAt = time.Now()
	}
	if person.UpdatedAt.IsZero() {
		person.UpdatedAt = time.Now()
	}

	return person, nil
}

// MapAttendee converts a TinyDB attendee to a domain Attendee
func (m *Mapper) MapAttendee(tdb *TinyDBAttendee, activityIDMap, personIDMap map[int]int64) (*domain.Attendee, error) {
	if tdb == nil {
		return nil, fmt.Errorf("tinydb attendee is nil")
	}

	// Map IDs from TinyDB to SQLite
	activityID, ok := activityIDMap[tdb.ActivityID]
	if !ok {
		return nil, fmt.Errorf("activity ID %d not found in mapping", tdb.ActivityID)
	}

	personID, ok := personIDMap[tdb.PersonID]
	if !ok {
		return nil, fmt.Errorf("person ID %d not found in mapping", tdb.PersonID)
	}

	// Map role
	role, err := m.mapAttendeeRole(tdb.Role)
	if err != nil {
		return nil, err
	}

	// Create attendee
	attendee := &domain.Attendee{
		ActivityID:    activityID,
		PersonID:      personID,
		Role:          role,
		PaymentAmount: decimal.NewFromFloat(tdb.PaymentAmount),
		Notes:         tdb.Notes,
	}

	// Map payment status
	paymentStatus, err := m.mapPaymentStatus(tdb.PaymentStatus)
	if err != nil {
		return nil, err
	}
	attendee.PaymentStatus = paymentStatus

	// Parse registration date
	if tdb.RegistrationDate != "" {
		regDate, err := parseDateTime(tdb.RegistrationDate)
		if err == nil {
			attendee.RegistrationDate = regDate
		} else {
			attendee.RegistrationDate = time.Now()
		}
	} else {
		attendee.RegistrationDate = time.Now()
	}

	// Parse payment date
	if tdb.PaymentDate != "" {
		paymentDate, err := parseDateTime(tdb.PaymentDate)
		if err == nil {
			attendee.PaymentDate = &paymentDate
		}
	}

	return attendee, nil
}

// mapActivityType maps TinyDB activity type string to domain ActivityType
func (m *Mapper) mapActivityType(typeStr string) (domain.ActivityType, error) {
	switch strings.ToLower(typeStr) {
	case "event":
		return domain.ActivityTypeEvent, nil
	case "gathering":
		return domain.ActivityTypeGathering, nil
	default:
		// Default to event if unknown
		return domain.ActivityTypeEvent, nil
	}
}

// mapActivityStatus maps TinyDB status string to domain ActivityStatus
func (m *Mapper) mapActivityStatus(statusStr string) (domain.ActivityStatus, error) {
	switch strings.ToLower(statusStr) {
	case "active":
		return domain.StatusActive, nil
	case "cancelled", "canceled":
		return domain.StatusCancelled, nil
	case "completed":
		return domain.StatusCompleted, nil
	default:
		// Default to active if unknown
		return domain.StatusActive, nil
	}
}

// mapAttendeeRole maps TinyDB role string to domain AttendeeRole
func (m *Mapper) mapAttendeeRole(roleStr string) (domain.AttendeeRole, error) {
	switch strings.ToLower(roleStr) {
	case "participant":
		return domain.RoleParticipant, nil
	case "volunteer":
		return domain.RoleVolunteer, nil
	case "worship_team", "worship team":
		return domain.RoleWorshipTeam, nil
	case "workshop_leader", "workshop leader":
		return domain.RoleWorkshopLeader, nil
	default:
		// Default to participant if unknown
		return domain.RoleParticipant, nil
	}
}

// mapPaymentStatus maps TinyDB payment status string to domain PaymentStatus
func (m *Mapper) mapPaymentStatus(statusStr string) (domain.PaymentStatus, error) {
	switch strings.ToLower(statusStr) {
	case "paid":
		return domain.PaymentStatusPaid, nil
	case "unpaid":
		return domain.PaymentStatusUnpaid, nil
	case "waived":
		return domain.PaymentStatusWaived, nil
	default:
		// Default to unpaid if unknown
		return domain.PaymentStatusUnpaid, nil
	}
}
