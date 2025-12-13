package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/PodioSpaz/event-tracker-go/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapacityService_GetCapacityInfo(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	ctx := context.Background()

	t.Run("unlimited capacity", func(t *testing.T) {
		// Create activity with no capacity limit
		activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
		require.NoError(t, activityRepo.Create(ctx, activity))

		info, err := capacitySvc.GetCapacityInfo(ctx, activity.ID)
		require.NoError(t, err)
		assert.Nil(t, info.MaxCapacity)
		assert.True(t, info.IsUnlimited)
		assert.Equal(t, -1, info.Remaining)
		assert.False(t, info.IsFull)
		assert.Equal(t, 0.0, info.PercentageFilled)
	})

	t.Run("limited capacity with attendees", func(t *testing.T) {
		// Create activity with capacity of 10
		activity := domain.NewActivity("Test Event 2", "Test Location", time.Now().AddDate(0, 1, 0))
		capacity := 10
		activity.SetCapacity(&capacity)
		require.NoError(t, activityRepo.Create(ctx, activity))

		// Register 3 attendees
		for i := 0; i < 3; i++ {
			person := domain.NewPerson("Person", string(rune('A'+i)), fmt.Sprintf("person%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", i))
			require.NoError(t, personRepo.Create(ctx, person))

			_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
			require.NoError(t, err)
		}

		info, err := capacitySvc.GetCapacityInfo(ctx, activity.ID)
		require.NoError(t, err)
		assert.NotNil(t, info.MaxCapacity)
		assert.Equal(t, 10, *info.MaxCapacity)
		assert.False(t, info.IsUnlimited)
		assert.Equal(t, 3, info.CurrentAttendees)
		assert.Equal(t, 7, info.Remaining)
		assert.False(t, info.IsFull)
		assert.Equal(t, 30.0, info.PercentageFilled)
	})

	t.Run("full capacity", func(t *testing.T) {
		// Create activity with capacity of 2
		activity := domain.NewActivity("Test Event 3", "Test Location", time.Now().AddDate(0, 1, 0))
		capacity := 2
		activity.SetCapacity(&capacity)
		require.NoError(t, activityRepo.Create(ctx, activity))

		// Register 2 attendees
		for i := 0; i < 2; i++ {
			person := domain.NewPerson("FullPerson", string(rune('A'+i)), fmt.Sprintf("fullperson%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", 1000+i))
			require.NoError(t, personRepo.Create(ctx, person))

			_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
			require.NoError(t, err)
		}

		info, err := capacitySvc.GetCapacityInfo(ctx, activity.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, info.CurrentAttendees)
		assert.Equal(t, 0, info.Remaining)
		assert.True(t, info.IsFull)
		assert.Equal(t, 100.0, info.PercentageFilled)
	})

	t.Run("activity not found", func(t *testing.T) {
		_, err := capacitySvc.GetCapacityInfo(ctx, 99999)
		require.Error(t, err)
	})
}

func TestCapacityService_CanAcceptAttendees(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	ctx := context.Background()

	// Create activity with capacity of 10
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity := 10
	activity.SetCapacity(&capacity)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Register 5 attendees
	for i := 0; i < 5; i++ {
		person := domain.NewPerson("Person", string(rune('A'+i)), fmt.Sprintf("person%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", i))
		require.NoError(t, personRepo.Create(ctx, person))

		_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
	}

	t.Run("can accept 3 more", func(t *testing.T) {
		canAccept, err := capacitySvc.CanAcceptAttendees(ctx, activity.ID, 3)
		require.NoError(t, err)
		assert.True(t, canAccept)
	})

	t.Run("can accept 5 more", func(t *testing.T) {
		canAccept, err := capacitySvc.CanAcceptAttendees(ctx, activity.ID, 5)
		require.NoError(t, err)
		assert.True(t, canAccept)
	})

	t.Run("cannot accept 6 more", func(t *testing.T) {
		canAccept, err := capacitySvc.CanAcceptAttendees(ctx, activity.ID, 6)
		require.NoError(t, err)
		assert.False(t, canAccept)
	})

	t.Run("unlimited capacity", func(t *testing.T) {
		// Create activity with unlimited capacity
		activity2 := domain.NewActivity("Unlimited Event", "Test Location", time.Now().AddDate(0, 1, 0))
		require.NoError(t, activityRepo.Create(ctx, activity2))

		canAccept, err := capacitySvc.CanAcceptAttendees(ctx, activity2.ID, 1000)
		require.NoError(t, err)
		assert.True(t, canAccept)
	})
}

func TestCapacityService_IsFull(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	ctx := context.Background()

	// Create activity with capacity of 2
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity := 2
	activity.SetCapacity(&capacity)
	require.NoError(t, activityRepo.Create(ctx, activity))

	t.Run("not full", func(t *testing.T) {
		isFull, err := capacitySvc.IsFull(ctx, activity.ID)
		require.NoError(t, err)
		assert.False(t, isFull)
	})

	// Register 2 attendees
	for i := 0; i < 2; i++ {
		person := domain.NewPerson("Person", string(rune('A'+i)), fmt.Sprintf("person%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", i))
		require.NoError(t, personRepo.Create(ctx, person))

		_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
	}

	t.Run("full", func(t *testing.T) {
		isFull, err := capacitySvc.IsFull(ctx, activity.ID)
		require.NoError(t, err)
		assert.True(t, isFull)
	})
}

func TestCapacityService_GetRemainingCapacity(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	ctx := context.Background()

	// Create activity with capacity of 10
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity := 10
	activity.SetCapacity(&capacity)
	require.NoError(t, activityRepo.Create(ctx, activity))

	t.Run("full capacity remaining", func(t *testing.T) {
		remaining, err := capacitySvc.GetRemainingCapacity(ctx, activity.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, remaining)
	})

	// Register 3 attendees
	for i := 0; i < 3; i++ {
		person := domain.NewPerson("Person", string(rune('A'+i)), fmt.Sprintf("person%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", i))
		require.NoError(t, personRepo.Create(ctx, person))

		_, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
	}

	t.Run("partial capacity remaining", func(t *testing.T) {
		remaining, err := capacitySvc.GetRemainingCapacity(ctx, activity.ID)
		require.NoError(t, err)
		assert.Equal(t, 7, remaining)
	})

	t.Run("unlimited capacity", func(t *testing.T) {
		activity2 := domain.NewActivity("Unlimited Event", "Test Location", time.Now().AddDate(0, 1, 0))
		require.NoError(t, activityRepo.Create(ctx, activity2))

		remaining, err := capacitySvc.GetRemainingCapacity(ctx, activity2.ID)
		require.NoError(t, err)
		assert.Equal(t, -1, remaining)
	})
}

func TestCapacityService_GetNearCapacityActivities(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	ctx := context.Background()

	// Create activity with capacity of 10 and 9 attendees (90% full)
	activity1 := domain.NewActivity("Near Capacity Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity1 := 10
	activity1.SetCapacity(&capacity1)
	require.NoError(t, activityRepo.Create(ctx, activity1))

	for i := 0; i < 9; i++ {
		person := domain.NewPerson("Person", string(rune('A'+i)), fmt.Sprintf("person%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", i))
		require.NoError(t, personRepo.Create(ctx, person))

		_, err := registrationSvc.RegisterAttendee(ctx, activity1.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
	}

	// Create activity with capacity of 20 and 5 attendees (25% full)
	activity2 := domain.NewActivity("Low Capacity Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity2 := 20
	activity2.SetCapacity(&capacity2)
	require.NoError(t, activityRepo.Create(ctx, activity2))

	for i := 0; i < 5; i++ {
		person := domain.NewPerson("LowPerson", string(rune('A'+i)), fmt.Sprintf("lowperson%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", 2000+i))
		require.NoError(t, personRepo.Create(ctx, person))

		_, err := registrationSvc.RegisterAttendee(ctx, activity2.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
	}

	t.Run("get activities above 80%", func(t *testing.T) {
		nearCapacity, err := capacitySvc.GetNearCapacityActivities(ctx, 80.0)
		require.NoError(t, err)
		assert.Len(t, nearCapacity, 1)
		assert.Equal(t, activity1.ID, nearCapacity[0].ID)
	})

	t.Run("get activities above 20%", func(t *testing.T) {
		nearCapacity, err := capacitySvc.GetNearCapacityActivities(ctx, 20.0)
		require.NoError(t, err)
		assert.Len(t, nearCapacity, 2)
	})
}

func TestCapacityService_GetFullActivities(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	capacitySvc := service.NewCapacityService(activityRepo, attendeeRepo)

	ctx := context.Background()

	// Create full activity
	activity1 := domain.NewActivity("Full Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity1 := 2
	activity1.SetCapacity(&capacity1)
	require.NoError(t, activityRepo.Create(ctx, activity1))

	for i := 0; i < 2; i++ {
		person := domain.NewPerson("FullPerson", string(rune('A'+i)), fmt.Sprintf("fullperson%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", 1000+i))
		require.NoError(t, personRepo.Create(ctx, person))

		_, err := registrationSvc.RegisterAttendee(ctx, activity1.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
	}

	// Create non-full activity
	activity2 := domain.NewActivity("Not Full Event", "Test Location", time.Now().AddDate(0, 1, 0))
	capacity2 := 10
	activity2.SetCapacity(&capacity2)
	require.NoError(t, activityRepo.Create(ctx, activity2))

	t.Run("get full activities", func(t *testing.T) {
		fullActivities, err := capacitySvc.GetFullActivities(ctx)
		require.NoError(t, err)
		assert.Len(t, fullActivities, 1)
		assert.Equal(t, activity1.ID, fullActivities[0].ID)
	})
}
