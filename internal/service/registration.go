package service

import (
	"context"
	"fmt"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// RegistrationService handles attendee registration logic
type RegistrationService struct {
	activityRepo repository.ActivityRepository
	personRepo   repository.PersonRepository
	attendeeRepo repository.AttendeeRepository
	transactor   repository.Transactor
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(
	activityRepo repository.ActivityRepository,
	personRepo repository.PersonRepository,
	attendeeRepo repository.AttendeeRepository,
	transactor repository.Transactor,
) *RegistrationService {
	return &RegistrationService{
		activityRepo: activityRepo,
		personRepo:   personRepo,
		attendeeRepo: attendeeRepo,
		transactor:   transactor,
	}
}

// RegisterAttendee registers a person for an activity
func (s *RegistrationService) RegisterAttendee(ctx context.Context, activityID, personID int64, role domain.AttendeeRole) (*domain.Attendee, error) {
	var attendee *domain.Attendee
	var err error

	// Execute registration within a transaction
	err = s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate that the person exists
		person, err := s.personRepo.GetByID(txCtx, personID)
		if err != nil {
			return fmt.Errorf("person not found: %w", err)
		}
		if person == nil {
			return util.ErrPersonNotFound
		}

		// Validate that the activity exists
		activity, err := s.activityRepo.GetByID(txCtx, activityID)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Check if already registered
		isRegistered, err := s.attendeeRepo.IsRegistered(txCtx, activityID, personID)
		if err != nil {
			return fmt.Errorf("failed to check registration: %w", err)
		}
		if isRegistered {
			return util.ErrAttendeeAlreadyRegistered
		}

		// Get current attendee count
		currentCount, err := s.attendeeRepo.CountByActivity(txCtx, activityID)
		if err != nil {
			return fmt.Errorf("failed to get attendee count: %w", err)
		}

		// Validate activity can accept registrations
		if err := activity.ValidateForRegistration(currentCount); err != nil {
			return err
		}

		// Create attendee
		attendee = domain.NewAttendee(activityID, personID, role)

		// Validate attendee
		if err := attendee.Validate(); err != nil {
			return err
		}

		// Create attendee record
		if err := s.attendeeRepo.Create(txCtx, attendee); err != nil {
			return fmt.Errorf("failed to create attendee: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return attendee, nil
}

// UnregisterAttendee removes an attendee registration
func (s *RegistrationService) UnregisterAttendee(ctx context.Context, attendeeID int64) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate that the attendee exists
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Delete attendee record
		if err := s.attendeeRepo.Delete(txCtx, attendeeID); err != nil {
			return fmt.Errorf("failed to delete attendee: %w", err)
		}

		return nil
	})
}

// UpdateAttendeeRole updates an attendee's role
func (s *RegistrationService) UpdateAttendeeRole(ctx context.Context, attendeeID int64, role domain.AttendeeRole) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get attendee
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Update role
		if err := attendee.SetRole(role); err != nil {
			return err
		}

		// Validate
		if err := attendee.Validate(); err != nil {
			return err
		}

		// Save
		if err := s.attendeeRepo.Update(txCtx, attendee); err != nil {
			return fmt.Errorf("failed to update attendee: %w", err)
		}

		return nil
	})
}

// GetAttendeesByActivity retrieves all attendees for an activity
func (s *RegistrationService) GetAttendeesByActivity(ctx context.Context, activityID int64) ([]*domain.Attendee, error) {
	// Validate activity exists
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}

	// Get attendees
	attendees, err := s.attendeeRepo.GetByActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendees: %w", err)
	}

	return attendees, nil
}

// GetAttendeesByPerson retrieves all registrations for a person
func (s *RegistrationService) GetAttendeesByPerson(ctx context.Context, personID int64) ([]*domain.Attendee, error) {
	// Validate person exists
	person, err := s.personRepo.GetByID(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("person not found: %w", err)
	}
	if person == nil {
		return nil, util.ErrPersonNotFound
	}

	// Get registrations
	attendees, err := s.attendeeRepo.GetByPerson(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registrations: %w", err)
	}

	return attendees, nil
}

// IsPersonRegistered checks if a person is registered for an activity
func (s *RegistrationService) IsPersonRegistered(ctx context.Context, activityID, personID int64) (bool, error) {
	return s.attendeeRepo.IsRegistered(ctx, activityID, personID)
}

// GetRegistrationCount returns the number of attendees for an activity
func (s *RegistrationService) GetRegistrationCount(ctx context.Context, activityID int64) (int, error) {
	return s.attendeeRepo.CountByActivity(ctx, activityID)
}

// AddAttendeeNotes adds or updates notes for an attendee
func (s *RegistrationService) AddAttendeeNotes(ctx context.Context, attendeeID int64, notes string) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get attendee
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Add notes
		attendee.AddNotes(notes)

		// Validate
		if err := attendee.Validate(); err != nil {
			return err
		}

		// Save
		if err := s.attendeeRepo.Update(txCtx, attendee); err != nil {
			return fmt.Errorf("failed to update attendee: %w", err)
		}

		return nil
	})
}
