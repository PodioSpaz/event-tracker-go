package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/service"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
	"github.com/shopspring/decimal"
)

// ActivityController coordinates activity-related operations
type ActivityController struct {
	activityRepo       repository.ActivityRepository
	registrationSvc    *service.RegistrationService
	paymentSvc         *service.PaymentService
	capacitySvc        *service.CapacityService
	transactor         repository.Transactor
}

// NewActivityController creates a new activity controller
func NewActivityController(
	activityRepo repository.ActivityRepository,
	registrationSvc *service.RegistrationService,
	paymentSvc *service.PaymentService,
	capacitySvc *service.CapacityService,
	transactor repository.Transactor,
) *ActivityController {
	return &ActivityController{
		activityRepo:    activityRepo,
		registrationSvc: registrationSvc,
		paymentSvc:      paymentSvc,
		capacitySvc:     capacitySvc,
		transactor:      transactor,
	}
}

// CreateActivity creates a new activity
func (c *ActivityController) CreateActivity(ctx context.Context, activity *domain.Activity) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate activity
		if err := activity.Validate(); err != nil {
			return err
		}

		// Create activity
		if err := c.activityRepo.Create(txCtx, activity); err != nil {
			return fmt.Errorf("failed to create activity: %w", err)
		}

		return nil
	})
}

// GetActivity retrieves an activity by ID
func (c *ActivityController) GetActivity(ctx context.Context, id int64) (*domain.Activity, error) {
	activity, err := c.activityRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}
	return activity, nil
}

// GetAllActivities retrieves all activities
func (c *ActivityController) GetAllActivities(ctx context.Context) ([]*domain.Activity, error) {
	activities, err := c.activityRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	return activities, nil
}

// UpdateActivity updates an existing activity
func (c *ActivityController) UpdateActivity(ctx context.Context, activity *domain.Activity) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate activity exists
		existing, err := c.activityRepo.GetByID(txCtx, activity.ID)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if existing == nil {
			return util.ErrActivityNotFound
		}

		// Validate updated activity
		if err := activity.Validate(); err != nil {
			return err
		}

		// Update timestamp
		activity.UpdatedAt = time.Now()

		// Update activity
		if err := c.activityRepo.Update(txCtx, activity); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		return nil
	})
}

// DeleteActivity deletes an activity
func (c *ActivityController) DeleteActivity(ctx context.Context, id int64) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Validate activity exists
		activity, err := c.activityRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Check if activity has attendees
		count, err := c.registrationSvc.GetRegistrationCount(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to check attendees: %w", err)
		}

		if count > 0 {
			return util.NewValidationError("activity", "cannot delete activity with attendees", util.ErrInvalidInput)
		}

		// Delete activity
		if err := c.activityRepo.Delete(txCtx, id); err != nil {
			return fmt.Errorf("failed to delete activity: %w", err)
		}

		return nil
	})
}

// CancelActivity marks an activity as cancelled
func (c *ActivityController) CancelActivity(ctx context.Context, id int64) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get activity
		activity, err := c.activityRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Cancel activity
		if err := activity.Cancel(); err != nil {
			return err
		}

		// Update activity
		if err := c.activityRepo.Update(txCtx, activity); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		return nil
	})
}

// CompleteActivity marks an activity as completed
func (c *ActivityController) CompleteActivity(ctx context.Context, id int64) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get activity
		activity, err := c.activityRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Complete activity
		if err := activity.Complete(); err != nil {
			return err
		}

		// Update activity
		if err := c.activityRepo.Update(txCtx, activity); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		return nil
	})
}

// ReactivateActivity marks a cancelled activity as active
func (c *ActivityController) ReactivateActivity(ctx context.Context, id int64) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get activity
		activity, err := c.activityRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Reactivate activity
		if err := activity.Reactivate(); err != nil {
			return err
		}

		// Update activity
		if err := c.activityRepo.Update(txCtx, activity); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		return nil
	})
}

// SetActivityFee sets the fee for an activity
func (c *ActivityController) SetActivityFee(ctx context.Context, id int64, fee decimal.Decimal) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get activity
		activity, err := c.activityRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Set fee
		if err := activity.SetFee(fee); err != nil {
			return err
		}

		// Validate
		if err := activity.Validate(); err != nil {
			return err
		}

		// Update activity
		if err := c.activityRepo.Update(txCtx, activity); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		return nil
	})
}

// SetActivityCapacity sets the maximum capacity for an activity
func (c *ActivityController) SetActivityCapacity(ctx context.Context, id int64, capacity *int) error {
	return c.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get activity
		activity, err := c.activityRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Check if reducing capacity below current attendee count
		if capacity != nil {
			currentCount, err := c.registrationSvc.GetRegistrationCount(txCtx, id)
			if err != nil {
				return fmt.Errorf("failed to get attendee count: %w", err)
			}

			if *capacity < currentCount {
				return util.NewValidationError("capacity",
					fmt.Sprintf("cannot set capacity below current attendee count (%d)", currentCount),
					util.ErrValueOutOfRange)
			}
		}

		// Set capacity
		if err := activity.SetCapacity(capacity); err != nil {
			return err
		}

		// Validate
		if err := activity.Validate(); err != nil {
			return err
		}

		// Update activity
		if err := c.activityRepo.Update(txCtx, activity); err != nil {
			return fmt.Errorf("failed to update activity: %w", err)
		}

		return nil
	})
}

// GetActivityWithCapacity retrieves an activity with capacity information
func (c *ActivityController) GetActivityWithCapacity(ctx context.Context, id int64) (*domain.Activity, *service.CapacityInfo, error) {
	// Get activity
	activity, err := c.GetActivity(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Get capacity info
	capacityInfo, err := c.capacitySvc.GetCapacityInfo(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get capacity info: %w", err)
	}

	return activity, capacityInfo, nil
}

// GetActivityWithPaymentSummary retrieves an activity with payment summary
func (c *ActivityController) GetActivityWithPaymentSummary(ctx context.Context, id int64) (*domain.Activity, *repository.PaymentSummary, error) {
	// Get activity
	activity, err := c.GetActivity(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Get payment summary
	summary, err := c.paymentSvc.GetPaymentSummary(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get payment summary: %w", err)
	}

	return activity, summary, nil
}

// GetUpcomingActivities retrieves upcoming activities
func (c *ActivityController) GetUpcomingActivities(ctx context.Context, limit int) ([]*domain.Activity, error) {
	activities, err := c.activityRepo.FindUpcoming(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming activities: %w", err)
	}
	return activities, nil
}

// GetActivitiesByStatus retrieves activities by status
func (c *ActivityController) GetActivitiesByStatus(ctx context.Context, status domain.ActivityStatus) ([]*domain.Activity, error) {
	activities, err := c.activityRepo.FindByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by status: %w", err)
	}
	return activities, nil
}

// GetActivitiesByType retrieves activities by type
func (c *ActivityController) GetActivitiesByType(ctx context.Context, activityType domain.ActivityType) ([]*domain.Activity, error) {
	activities, err := c.activityRepo.FindByType(ctx, activityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by type: %w", err)
	}
	return activities, nil
}

// GetActivitiesByDateRange retrieves activities within a date range
func (c *ActivityController) GetActivitiesByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Activity, error) {
	activities, err := c.activityRepo.FindByDateRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by date range: %w", err)
	}
	return activities, nil
}

// RegisterAttendee registers a person for an activity
func (c *ActivityController) RegisterAttendee(ctx context.Context, activityID, personID int64, role domain.AttendeeRole) (*domain.Attendee, error) {
	return c.registrationSvc.RegisterAttendee(ctx, activityID, personID, role)
}

// UnregisterAttendee removes an attendee registration
func (c *ActivityController) UnregisterAttendee(ctx context.Context, attendeeID int64) error {
	return c.registrationSvc.UnregisterAttendee(ctx, attendeeID)
}

// GetAttendees retrieves all attendees for an activity
func (c *ActivityController) GetAttendees(ctx context.Context, activityID int64) ([]*domain.Attendee, error) {
	return c.registrationSvc.GetAttendeesByActivity(ctx, activityID)
}

// MarkAttendeePaid marks an attendee as paid
func (c *ActivityController) MarkAttendeePaid(ctx context.Context, attendeeID int64, amount decimal.Decimal) error {
	return c.paymentSvc.MarkPaidNow(ctx, attendeeID, amount)
}

// WaiveAttendeePayment waives payment for an attendee
func (c *ActivityController) WaiveAttendeePayment(ctx context.Context, attendeeID int64) error {
	return c.paymentSvc.WaivePaymentNow(ctx, attendeeID)
}

// BulkMarkAttendeesPaid marks multiple attendees as paid
func (c *ActivityController) BulkMarkAttendeesPaid(ctx context.Context, attendeeIDs []int64, amount decimal.Decimal) error {
	return c.paymentSvc.BulkMarkPaid(ctx, attendeeIDs, amount, time.Now())
}

// BulkWaiveAttendeesPayment waives payment for multiple attendees
func (c *ActivityController) BulkWaiveAttendeesPayment(ctx context.Context, attendeeIDs []int64) error {
	return c.paymentSvc.BulkWaivePayment(ctx, attendeeIDs, time.Now())
}

// GetPaymentSummary retrieves payment summary for an activity
func (c *ActivityController) GetPaymentSummary(ctx context.Context, activityID int64) (*repository.PaymentSummary, error) {
	return c.paymentSvc.GetPaymentSummary(ctx, activityID)
}
