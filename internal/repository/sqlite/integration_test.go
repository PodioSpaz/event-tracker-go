package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sqlite.DB {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sqlite.New(dbPath)
	require.NoError(t, err)

	// Run migrations
	migrationsPath := "../../../migrations"
	err = db.RunMigrations(migrationsPath)
	require.NoError(t, err)

	return db
}

func TestActivityRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewActivityRepository(db)
	ctx := context.Background()

	// Test Create
	activity := domain.NewActivity(
		"Test Event",
		"Test Location",
		time.Now().AddDate(0, 0, 7),
	)
	activity.Description = "Test Description"
	activity.Fee = decimal.NewFromFloat(25.50)

	err := repo.Create(ctx, activity)
	assert.NoError(t, err)
	assert.NotZero(t, activity.ID)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, activity.ID)
	assert.NoError(t, err)
	assert.Equal(t, activity.Name, retrieved.Name)
	assert.Equal(t, activity.Location, retrieved.Location)
	assert.Equal(t, activity.Fee.String(), retrieved.Fee.String())

	// Test Update
	retrieved.Description = "Updated Description"
	err = repo.Update(ctx, retrieved)
	assert.NoError(t, err)

	updated, err := repo.GetByID(ctx, activity.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Description", updated.Description)

	// Test FindByStatus
	activities, err := repo.FindByStatus(ctx, domain.StatusActive)
	assert.NoError(t, err)
	assert.Len(t, activities, 1)

	// Test Count
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test Delete
	err = repo.Delete(ctx, activity.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = repo.GetByID(ctx, activity.ID)
	assert.Error(t, err)
}

func TestPersonRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewPersonRepository(db)
	ctx := context.Background()

	// Test Create
	person := domain.NewPerson("John", "Doe", "john.doe@example.com", "5551234567")

	err := repo.Create(ctx, person)
	assert.NoError(t, err)
	assert.NotZero(t, person.ID)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, person.ID)
	assert.NoError(t, err)
	assert.Equal(t, person.FirstName, retrieved.FirstName)
	assert.Equal(t, person.LastName, retrieved.LastName)
	assert.Equal(t, person.Email, retrieved.Email)

	// Test FindByEmail
	found, err := repo.FindByEmail(ctx, "john.doe@example.com")
	assert.NoError(t, err)
	assert.Equal(t, person.ID, found.ID)

	// Test Search
	results, err := repo.Search(ctx, "John")
	assert.NoError(t, err)
	assert.Len(t, results, 1)

	// Test duplicate email prevention
	duplicate := domain.NewPerson("Jane", "Doe", "john.doe@example.com", "5559999999")
	err = repo.Create(ctx, duplicate)
	assert.Error(t, err)

	// Test Update
	retrieved.Phone = "5555555555"
	err = repo.Update(ctx, retrieved)
	assert.NoError(t, err)

	updated, err := repo.GetByID(ctx, person.ID)
	assert.NoError(t, err)
	assert.Equal(t, "5555555555", updated.Phone)

	// Test Count
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test Delete
	err = repo.Delete(ctx, person.ID)
	assert.NoError(t, err)
}

func TestAttendeeRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 0, 7))
	activity.Fee = decimal.NewFromFloat(25.50)
	err := activityRepo.Create(ctx, activity)
	require.NoError(t, err)

	// Create test person
	person := domain.NewPerson("John", "Doe", "john@example.com", "")
	err = personRepo.Create(ctx, person)
	require.NoError(t, err)

	// Test Create attendee
	attendee := domain.NewAttendee(activity.ID, person.ID, domain.RoleParticipant)
	err = attendeeRepo.Create(ctx, attendee)
	assert.NoError(t, err)
	assert.NotZero(t, attendee.ID)

	// Test duplicate registration prevention
	duplicate := domain.NewAttendee(activity.ID, person.ID, domain.RoleVolunteer)
	err = attendeeRepo.Create(ctx, duplicate)
	assert.Error(t, err)

	// Test IsRegistered
	registered, err := attendeeRepo.IsRegistered(ctx, activity.ID, person.ID)
	assert.NoError(t, err)
	assert.True(t, registered)

	// Test GetByActivity
	attendees, err := attendeeRepo.GetByActivity(ctx, activity.ID)
	assert.NoError(t, err)
	assert.Len(t, attendees, 1)

	// Test mark paid
	retrieved, err := attendeeRepo.GetByID(ctx, attendee.ID)
	assert.NoError(t, err)

	err = retrieved.MarkPaidNow(decimal.NewFromFloat(25.50))
	assert.NoError(t, err)

	err = attendeeRepo.Update(ctx, retrieved)
	assert.NoError(t, err)

	// Test GetPaymentSummary
	summary, err := attendeeRepo.GetPaymentSummary(ctx, activity.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, summary.TotalAttendees)
	assert.Equal(t, 1, summary.PaidCount)
	assert.Equal(t, "25.5", summary.PaidAmount.String())

	// Test FindByPaymentStatus
	paid, err := attendeeRepo.FindByPaymentStatus(ctx, activity.ID, domain.PaymentStatusPaid)
	assert.NoError(t, err)
	assert.Len(t, paid, 1)

	// Test CountByActivity
	count, err := attendeeRepo.CountByActivity(ctx, activity.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test Delete
	err = attendeeRepo.Delete(ctx, attendee.ID)
	assert.NoError(t, err)
}

func TestRoleRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := sqlite.NewRoleRepository(db)
	ctx := context.Background()

	// Test GetAll (should have default roles)
	roles, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(roles), 4) // At least 4 default roles

	// Test GetByName
	participantRole, err := repo.GetByName(ctx, "participant")
	assert.NoError(t, err)
	assert.Equal(t, "Participant", participantRole.DisplayName)

	// Test GetActive
	activeRoles, err := repo.GetActive(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeRoles), 4)

	// Test Create
	customRole := domain.NewRole("custom_role", "Custom Role", "A custom role for testing")
	err = repo.Create(ctx, customRole)
	assert.NoError(t, err)
	assert.NotZero(t, customRole.ID)

	// Test duplicate name prevention
	duplicate := domain.NewRole("custom_role", "Duplicate", "Should fail")
	err = repo.Create(ctx, duplicate)
	assert.Error(t, err)

	// Test Update
	customRole.Description = "Updated description"
	err = repo.Update(ctx, customRole)
	assert.NoError(t, err)

	updated, err := repo.GetByID(ctx, customRole.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description)

	// Test Deactivate
	customRole.Deactivate()
	err = repo.Update(ctx, customRole)
	assert.NoError(t, err)

	// Verify not in active list
	activeRoles, err = repo.GetActive(ctx)
	assert.NoError(t, err)
	for _, role := range activeRoles {
		assert.NotEqual(t, customRole.ID, role.ID)
	}

	// Test Delete
	err = repo.Delete(ctx, customRole.ID)
	assert.NoError(t, err)
}

func TestTransaction_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	ctx := context.Background()

	// Test successful transaction
	err := db.WithTransaction(ctx, func(ctx context.Context) error {
		// Create activity
		activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 0, 7))
		if err := activityRepo.Create(ctx, activity); err != nil {
			return err
		}

		// Create person
		person := domain.NewPerson("John", "Doe", "john@example.com", "")
		if err := personRepo.Create(ctx, person); err != nil {
			return err
		}

		return nil
	})
	assert.NoError(t, err)

	// Verify both were created
	activities, err := activityRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, activities, 1)

	people, err := personRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, people, 1)

	// Test rollback on error
	err = db.WithTransaction(ctx, func(ctx context.Context) error {
		// Create activity
		activity := domain.NewActivity("Another Event", "Another Location", time.Now().AddDate(0, 0, 14))
		if err := activityRepo.Create(ctx, activity); err != nil {
			return err
		}

		// Return error to trigger rollback
		return assert.AnError
	})
	assert.Error(t, err)

	// Verify activity was rolled back
	activities, err = activityRepo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, activities, 1) // Still only 1 from successful transaction
}
