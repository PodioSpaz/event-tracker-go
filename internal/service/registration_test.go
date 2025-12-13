package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/PodioSpaz/event-tracker-go/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationService_RegisterAttendee(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	activity.SetFee(decimal.NewFromFloat(50.00))
	capacity := 10
	activity.SetCapacity(&capacity)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@example.com", "555-1234")
	require.NoError(t, personRepo.Create(ctx, person))

	t.Run("successful registration", func(t *testing.T) {
		attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
		assert.NotNil(t, attendee)
		assert.Equal(t, activity.ID, attendee.ActivityID)
		assert.Equal(t, person.ID, attendee.PersonID)
		assert.Equal(t, domain.RoleParticipant, attendee.Role)
		assert.Equal(t, domain.PaymentStatusUnpaid, attendee.PaymentStatus)
	})

	t.Run("duplicate registration", func(t *testing.T) {
		_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
		require.Error(t, err)
	})

	t.Run("activity not found", func(t *testing.T) {
		_, err := registrationSvc.RegisterAttendee(ctx, 99999, person.ID, domain.RoleParticipant)
		require.Error(t, err)
	})

	t.Run("person not found", func(t *testing.T) {
		_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, 99999, domain.RoleParticipant)
		require.Error(t, err)
	})
}

func TestRegistrationService_RegisterAttendee_CapacityValidation(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity with capacity of 2
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity := 2
	activity.SetCapacity(&capacity)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test people
	person1 := domain.NewPerson("John", "Doe", "john@example.com", "555-0001")
	require.NoError(t, personRepo.Create(ctx, person1))

	person2 := domain.NewPerson("Jane", "Doe", "jane@example.com", "555-0002")
	require.NoError(t, personRepo.Create(ctx, person2))

	person3 := domain.NewPerson("Bob", "Smith", "bob@example.com", "555-0003")
	require.NoError(t, personRepo.Create(ctx, person3))

	// Register first two attendees
	_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person1.ID, domain.RoleParticipant)
	require.NoError(t, err)

	_, err = registrationSvc.RegisterAttendee(ctx, activity.ID, person2.ID, domain.RoleParticipant)
	require.NoError(t, err)

	// Third registration should fail due to capacity
	_, err = registrationSvc.RegisterAttendee(ctx, activity.ID, person3.ID, domain.RoleParticipant)
	require.Error(t, err)
}

func TestRegistrationService_UnregisterAttendee(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person))

	// Register attendee
	attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
	require.NoError(t, err)

	t.Run("successful unregister", func(t *testing.T) {
		err := registrationSvc.UnregisterAttendee(ctx, attendee.ID)
		require.NoError(t, err)

		// Verify attendee is deleted
		_, err = attendeeRepo.GetByID(ctx, attendee.ID)
		require.Error(t, err)
	})

	t.Run("attendee not found", func(t *testing.T) {
		err := registrationSvc.UnregisterAttendee(ctx, 99999)
		require.Error(t, err)
	})
}

func TestRegistrationService_UpdateAttendeeRole(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person))

	// Register attendee
	attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
	require.NoError(t, err)

	t.Run("successful role update", func(t *testing.T) {
		err := registrationSvc.UpdateAttendeeRole(ctx, attendee.ID, domain.RoleVolunteer)
		require.NoError(t, err)

		// Verify role was updated
		updated, err := attendeeRepo.GetByID(ctx, attendee.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.RoleVolunteer, updated.Role)
	})

	t.Run("invalid role", func(t *testing.T) {
		err := registrationSvc.UpdateAttendeeRole(ctx, attendee.ID, domain.AttendeeRole("invalid"))
		require.Error(t, err)
	})

	t.Run("attendee not found", func(t *testing.T) {
		err := registrationSvc.UpdateAttendeeRole(ctx, 99999, domain.RoleParticipant)
		require.Error(t, err)
	})
}

func TestRegistrationService_GetAttendeesByActivity(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test people
	person1 := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person1))

	person2 := domain.NewPerson("Jane", "Doe", "jane.doe@test.com", "555-9002")
	require.NoError(t, personRepo.Create(ctx, person2))

	// Register attendees
	_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person1.ID, domain.RoleParticipant)
	require.NoError(t, err)

	_, err = registrationSvc.RegisterAttendee(ctx, activity.ID, person2.ID, domain.RoleVolunteer)
	require.NoError(t, err)

	t.Run("get all attendees", func(t *testing.T) {
		attendees, err := registrationSvc.GetAttendeesByActivity(ctx, activity.ID)
		require.NoError(t, err)
		assert.Len(t, attendees, 2)
	})

	t.Run("activity not found", func(t *testing.T) {
		_, err := registrationSvc.GetAttendeesByActivity(ctx, 99999)
		require.Error(t, err)
	})
}

func TestRegistrationService_AddAttendeeNotes(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person))

	// Register attendee
	attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
	require.NoError(t, err)

	t.Run("add notes", func(t *testing.T) {
		notes := "Special dietary requirements"
		err := registrationSvc.AddAttendeeNotes(ctx, attendee.ID, notes)
		require.NoError(t, err)

		// Verify notes were added
		updated, err := attendeeRepo.GetByID(ctx, attendee.ID)
		require.NoError(t, err)
		assert.Equal(t, notes, updated.Notes)
	})
}

// setupTestDB creates a test database and returns cleanup function
func setupTestDB(t *testing.T) (*sqlite.DB, func()) {
	db, err := sqlite.New(":memory:")
	require.NoError(t, err)

	// Run migrations
	err = db.RunMigrations("../../migrations")
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
