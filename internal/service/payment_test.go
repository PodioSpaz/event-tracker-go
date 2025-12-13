package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository/sqlite"
	"github.com/PodioSpaz/event-tracker-go/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentService_MarkPaid(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	fee := decimal.NewFromFloat(50.00)
	activity.SetFee(fee)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person))

	// Register attendee
	attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
	require.NoError(t, err)

	t.Run("successful payment", func(t *testing.T) {
		paymentDate := time.Now()
		err := paymentSvc.MarkPaid(ctx, attendee.ID, fee, paymentDate)
		require.NoError(t, err)

		// Verify payment status
		updated, err := attendeeRepo.GetByID(ctx, attendee.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusPaid, updated.PaymentStatus)
		assert.True(t, updated.PaymentAmount.Equal(fee))
		assert.NotNil(t, updated.PaymentDate)
	})

	t.Run("payment amount mismatch", func(t *testing.T) {
		// Create new attendee for this test
		person2 := domain.NewPerson("Jane", "Doe", "jane.doe@test.com", "555-9002")
		require.NoError(t, personRepo.Create(ctx, person2))

		attendee2, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person2.ID, domain.RoleParticipant)
		require.NoError(t, err)

		wrongAmount := decimal.NewFromFloat(25.00)
		err = paymentSvc.MarkPaidNow(ctx, attendee2.ID, wrongAmount)
		require.Error(t, err)
	})

	t.Run("negative payment amount", func(t *testing.T) {
		// Create new attendee for this test
		person3 := domain.NewPerson("Bob", "Smith", "bob.smith@test.com", "555-9003")
		require.NoError(t, personRepo.Create(ctx, person3))

		attendee3, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person3.ID, domain.RoleParticipant)
		require.NoError(t, err)

		negativeAmount := decimal.NewFromFloat(-50.00)
		err = paymentSvc.MarkPaidNow(ctx, attendee3.ID, negativeAmount)
		require.Error(t, err)
	})

	t.Run("attendee not found", func(t *testing.T) {
		err := paymentSvc.MarkPaidNow(ctx, 99999, fee)
		require.Error(t, err)
	})
}

func TestPaymentService_MarkUnpaid(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	fee := decimal.NewFromFloat(50.00)
	activity.SetFee(fee)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person))

	// Register attendee and mark as paid
	attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
	require.NoError(t, err)

	err = paymentSvc.MarkPaidNow(ctx, attendee.ID, fee)
	require.NoError(t, err)

	t.Run("mark unpaid", func(t *testing.T) {
		err := paymentSvc.MarkUnpaid(ctx, attendee.ID)
		require.NoError(t, err)

		// Verify payment status
		updated, err := attendeeRepo.GetByID(ctx, attendee.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusUnpaid, updated.PaymentStatus)
		assert.True(t, updated.PaymentAmount.IsZero())
		assert.Nil(t, updated.PaymentDate)
	})
}

func TestPaymentService_WaivePayment(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	fee := decimal.NewFromFloat(50.00)
	activity.SetFee(fee)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test person
	person := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person))

	// Register attendee
	attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
	require.NoError(t, err)

	t.Run("waive payment", func(t *testing.T) {
		waiveDate := time.Now()
		err := paymentSvc.WaivePayment(ctx, attendee.ID, waiveDate)
		require.NoError(t, err)

		// Verify payment status
		updated, err := attendeeRepo.GetByID(ctx, attendee.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusWaived, updated.PaymentStatus)
		assert.True(t, updated.PaymentAmount.IsZero())
		assert.NotNil(t, updated.PaymentDate)
	})

	t.Run("waive payment now", func(t *testing.T) {
		// Create new attendee
		person2 := domain.NewPerson("Jane", "Doe", "jane.doe@test.com", "555-9002")
		require.NoError(t, personRepo.Create(ctx, person2))

		attendee2, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person2.ID, domain.RoleParticipant)
		require.NoError(t, err)

		err = paymentSvc.WaivePaymentNow(ctx, attendee2.ID)
		require.NoError(t, err)

		// Verify payment status
		updated, err := attendeeRepo.GetByID(ctx, attendee2.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.PaymentStatusWaived, updated.PaymentStatus)
	})
}

func TestPaymentService_GetPaymentSummary(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	fee := decimal.NewFromFloat(50.00)
	activity.SetFee(fee)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test people
	person1 := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person1))

	person2 := domain.NewPerson("Jane", "Doe", "jane.doe@test.com", "555-9002")
	require.NoError(t, personRepo.Create(ctx, person2))

	person3 := domain.NewPerson("Bob", "Smith", "bob.smith@test.com", "555-9003")
	require.NoError(t, personRepo.Create(ctx, person3))

	// Register attendees
	attendee1, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person1.ID, domain.RoleParticipant)
	require.NoError(t, err)

	attendee2, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person2.ID, domain.RoleParticipant)
	require.NoError(t, err)

	_, err = registrationSvc.RegisterAttendee(ctx, activity.ID, person3.ID, domain.RoleParticipant)
	require.NoError(t, err)

	// Mark payments
	err = paymentSvc.MarkPaidNow(ctx, attendee1.ID, fee)
	require.NoError(t, err)

	err = paymentSvc.WaivePaymentNow(ctx, attendee2.ID)
	require.NoError(t, err)
	// attendee3 remains unpaid

	t.Run("get payment summary", func(t *testing.T) {
		summary, err := paymentSvc.GetPaymentSummary(ctx, activity.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, summary.TotalAttendees)
		assert.Equal(t, 1, summary.PaidCount)
		assert.Equal(t, 1, summary.UnpaidCount)
		assert.Equal(t, 1, summary.WaivedCount)
	})
}

func TestPaymentService_GetAttendeesByPaymentStatus(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	fee := decimal.NewFromFloat(50.00)
	activity.SetFee(fee)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test people
	person1 := domain.NewPerson("John", "Doe", "john.doe@test.com", "555-9001")
	require.NoError(t, personRepo.Create(ctx, person1))

	person2 := domain.NewPerson("Jane", "Doe", "jane.doe@test.com", "555-9002")
	require.NoError(t, personRepo.Create(ctx, person2))

	person3 := domain.NewPerson("Bob", "Smith", "bob.smith@test.com", "555-9003")
	require.NoError(t, personRepo.Create(ctx, person3))

	// Register attendees
	attendee1, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person1.ID, domain.RoleParticipant)
	require.NoError(t, err)

	attendee2, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person2.ID, domain.RoleParticipant)
	require.NoError(t, err)

	attendee3, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person3.ID, domain.RoleParticipant)
	require.NoError(t, err)

	// Mark payments
	err = paymentSvc.MarkPaidNow(ctx, attendee1.ID, fee)
	require.NoError(t, err)

	err = paymentSvc.WaivePaymentNow(ctx, attendee2.ID)
	require.NoError(t, err)
	// attendee3 remains unpaid

	t.Run("get paid attendees", func(t *testing.T) {
		paid, err := paymentSvc.GetPaidAttendees(ctx, activity.ID)
		require.NoError(t, err)
		assert.Len(t, paid, 1)
		assert.Equal(t, attendee1.ID, paid[0].ID)
	})

	t.Run("get unpaid attendees", func(t *testing.T) {
		unpaid, err := paymentSvc.GetUnpaidAttendees(ctx, activity.ID)
		require.NoError(t, err)
		assert.Len(t, unpaid, 1)
		assert.Equal(t, attendee3.ID, unpaid[0].ID)
	})

	t.Run("get waived attendees", func(t *testing.T) {
		waived, err := paymentSvc.GetWaivedAttendees(ctx, activity.ID)
		require.NoError(t, err)
		assert.Len(t, waived, 1)
		assert.Equal(t, attendee2.ID, waived[0].ID)
	})
}

func TestPaymentService_BulkOperations(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	activityRepo := sqlite.NewActivityRepository(db)
	personRepo := sqlite.NewPersonRepository(db)
	attendeeRepo := sqlite.NewAttendeeRepository(db)
	// DB implements Transactor interface

	registrationSvc := service.NewRegistrationService(activityRepo, personRepo, attendeeRepo, db)
	paymentSvc := service.NewPaymentService(activityRepo, attendeeRepo, db)

	ctx := context.Background()

	// Create test activity
	activity := domain.NewActivity("Test Event", "Test Location", time.Now().AddDate(0, 1, 0))
	fee := decimal.NewFromFloat(50.00)
	activity.SetFee(fee)
	require.NoError(t, activityRepo.Create(ctx, activity))

	// Create test people and register
	var attendeeIDs []int64
	for i := 0; i < 3; i++ {
		person := domain.NewPerson("Person", string(rune('A'+i)), fmt.Sprintf("person%c@example.com", rune('A'+i)), fmt.Sprintf("555-%04d", i))
		require.NoError(t, personRepo.Create(ctx, person))

		attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
		require.NoError(t, err)
		attendeeIDs = append(attendeeIDs, attendee.ID)
	}

	t.Run("bulk mark paid", func(t *testing.T) {
		err := paymentSvc.BulkMarkPaid(ctx, attendeeIDs, fee, time.Now())
		require.NoError(t, err)

		// Verify all are paid
		paid, err := paymentSvc.GetPaidAttendees(ctx, activity.ID)
		require.NoError(t, err)
		assert.Len(t, paid, 3)
	})

	t.Run("bulk waive payment", func(t *testing.T) {
		// Create new attendees
		var newAttendeeIDs []int64
		for i := 0; i < 2; i++ {
			person := domain.NewPerson("NewPerson", string(rune('X'+i)), fmt.Sprintf("newperson%c@example.com", rune('X'+i)), fmt.Sprintf("555-%04d", 3000+i))
			require.NoError(t, personRepo.Create(ctx, person))

			attendee, err := registrationSvc.RegisterAttendee(ctx, activity.ID, person.ID, domain.RoleParticipant)
			require.NoError(t, err)
			newAttendeeIDs = append(newAttendeeIDs, attendee.ID)
		}

		err := paymentSvc.BulkWaivePayment(ctx, newAttendeeIDs, time.Now())
		require.NoError(t, err)

		// Verify all are waived
		waived, err := paymentSvc.GetWaivedAttendees(ctx, activity.ID)
		require.NoError(t, err)
		assert.Len(t, waived, 2)
	})
}
