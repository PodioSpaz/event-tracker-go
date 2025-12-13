package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/service"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// PersonController coordinates person-related operations
type PersonController struct {
	personRepo      repository.PersonRepository
	registrationSvc *service.RegistrationService
	transactor      repository.Transactor
}

// NewPersonController creates a new person controller
func NewPersonController(
	personRepo repository.PersonRepository,
	registrationSvc *service.RegistrationService,
	transactor repository.Transactor,
) *PersonController {
	return &PersonController{
		personRepo:      personRepo,
		registrationSvc: registrationSvc,
		transactor:      transactor,
	}
}

// CreatePerson creates a new person
func (c *PersonController) CreatePerson(ctx context.Context, person *domain.Person) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate person
		if err := person.Validate(); err != nil {
			return err
		}

		// Check if email already exists
		if person.Email != "" {
			exists, err := c.personRepo.ExistsByEmail(txCtx, person.Email)
			if err != nil {
				return fmt.Errorf("failed to check email: %w", err)
			}
			if exists {
				return util.ErrPersonDuplicate
			}
		}

		// Set timestamps
		now := time.Now()
		person.CreatedAt = now
		person.UpdatedAt = now

		// Create person
		if err := c.personRepo.Create(txCtx, person); err != nil {
			return fmt.Errorf("failed to create person: %w", err)
		}

		return nil
	})
}

// GetPerson retrieves a person by ID
func (c *PersonController) GetPerson(ctx context.Context, id int64) (*domain.Person, error) {
	person, err := c.personRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("person not found: %w", err)
	}
	if person == nil {
		return nil, util.ErrPersonNotFound
	}
	return person, nil
}

// GetAllPeople retrieves all people
func (c *PersonController) GetAllPeople(ctx context.Context) ([]*domain.Person, error) {
	people, err := c.personRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get people: %w", err)
	}
	return people, nil
}

// UpdatePerson updates an existing person
func (c *PersonController) UpdatePerson(ctx context.Context, person *domain.Person) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate person exists
		existing, err := c.personRepo.GetByID(txCtx, person.ID)
		if err != nil {
			return fmt.Errorf("person not found: %w", err)
		}
		if existing == nil {
			return util.ErrPersonNotFound
		}

		// Validate updated person
		if err := person.Validate(); err != nil {
			return err
		}

		// Check if email changed and new email already exists
		if person.Email != "" && person.Email != existing.Email {
			exists, err := c.personRepo.ExistsByEmail(txCtx, person.Email)
			if err != nil {
				return fmt.Errorf("failed to check email: %w", err)
			}
			if exists {
				return util.ErrPersonDuplicate
			}
		}

		// Update timestamp
		person.UpdatedAt = time.Now()

		// Update person
		if err := c.personRepo.Update(txCtx, person); err != nil {
			return fmt.Errorf("failed to update person: %w", err)
		}

		return nil
	})
}

// DeletePerson deletes a person
func (c *PersonController) DeletePerson(ctx context.Context, id int64) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate person exists
		person, err := c.personRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("person not found: %w", err)
		}
		if person == nil {
			return util.ErrPersonNotFound
		}

		// Check if person has registrations
		registrations, err := c.registrationSvc.GetAttendeesByPerson(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to check registrations: %w", err)
		}

		if len(registrations) > 0 {
			return util.NewValidationError("person", "cannot delete person with registrations", util.ErrInvalidInput)
		}

		// Delete person
		if err := c.personRepo.Delete(txCtx, id); err != nil {
			return fmt.Errorf("failed to delete person: %w", err)
		}

		return nil
	})
}

// FindPersonByEmail retrieves a person by email
func (c *PersonController) FindPersonByEmail(ctx context.Context, email string) (*domain.Person, error) {
	person, err := c.personRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("person not found: %w", err)
	}
	if person == nil {
		return nil, util.ErrPersonNotFound
	}
	return person, nil
}

// FindPeopleByName searches for people by name
func (c *PersonController) FindPeopleByName(ctx context.Context, name string) ([]*domain.Person, error) {
	people, err := c.personRepo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search people: %w", err)
	}
	return people, nil
}

// SearchPeople searches for people by query (name or email)
func (c *PersonController) SearchPeople(ctx context.Context, query string) ([]*domain.Person, error) {
	people, err := c.personRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search people: %w", err)
	}
	return people, nil
}

// GetPersonRegistrations retrieves all activity registrations for a person
func (c *PersonController) GetPersonRegistrations(ctx context.Context, personID int64) ([]*domain.Attendee, error) {
	// Validate person exists
	_, err := c.GetPerson(ctx, personID)
	if err != nil {
		return nil, err
	}

	// Get registrations
	return c.registrationSvc.GetAttendeesByPerson(ctx, personID)
}

// GetPersonCount returns the total number of people
func (c *PersonController) GetPersonCount(ctx context.Context) (int, error) {
	count, err := c.personRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get person count: %w", err)
	}
	return count, nil
}

// Note: UpdatePersonNotes and TogglePersonActive are not implemented
// as the Person model does not currently have Notes and IsActive fields.
// These can be added when those fields are added to the domain model.

// BulkCreatePeople creates multiple people (useful for CSV import)
func (c *PersonController) BulkCreatePeople(ctx context.Context, people []*domain.Person) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, person := range people {
			// Validate person
			if err := person.Validate(); err != nil {
				return fmt.Errorf("validation failed for %s %s: %w", person.FirstName, person.LastName, err)
			}

			// Check if email already exists
			if person.Email != "" {
				exists, err := c.personRepo.ExistsByEmail(txCtx, person.Email)
				if err != nil {
					return fmt.Errorf("failed to check email for %s: %w", person.Email, err)
				}
				if exists {
					return fmt.Errorf("email already exists: %s", person.Email)
				}
			}

			// Set timestamps
			now := time.Now()
			person.CreatedAt = now
			person.UpdatedAt = now

			// Create person
			if err := c.personRepo.Create(txCtx, person); err != nil {
				return fmt.Errorf("failed to create person %s %s: %w", person.FirstName, person.LastName, err)
			}
		}

		return nil
	})
}
