package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
	"github.com/shopspring/decimal"
)

// PaymentService handles payment operations
type PaymentService struct {
	activityRepo repository.ActivityRepository
	attendeeRepo repository.AttendeeRepository
	transactor   repository.Transactor
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	activityRepo repository.ActivityRepository,
	attendeeRepo repository.AttendeeRepository,
	transactor repository.Transactor,
) *PaymentService {
	return &PaymentService{
		activityRepo: activityRepo,
		attendeeRepo: attendeeRepo,
		transactor:   transactor,
	}
}

// MarkPaid marks an attendee as paid
func (s *PaymentService) MarkPaid(ctx context.Context, attendeeID int64, amount decimal.Decimal, paymentDate time.Time) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get attendee
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Get activity to validate fee
		activity, err := s.activityRepo.GetByID(txCtx, attendee.ActivityID)
		if err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}
		if activity == nil {
			return util.ErrActivityNotFound
		}

		// Validate amount matches activity fee (unless it's a free event)
		if !activity.IsFree && !amount.Equal(activity.Fee) {
			return util.ErrPaymentAmountMismatch
		}

		// Mark as paid
		if err := attendee.MarkPaid(amount, paymentDate); err != nil {
			return err
		}

		// Validate
		if err := attendee.Validate(); err != nil {
			return err
		}

		// Validate payment amount against activity fee
		if err := attendee.ValidatePaymentAmount(activity.Fee); err != nil {
			return err
		}

		// Save
		if err := s.attendeeRepo.Update(txCtx, attendee); err != nil {
			return fmt.Errorf("failed to update attendee: %w", err)
		}

		return nil
	})
}

// MarkPaidNow marks an attendee as paid with the current timestamp
func (s *PaymentService) MarkPaidNow(ctx context.Context, attendeeID int64, amount decimal.Decimal) error {
	return s.MarkPaid(ctx, attendeeID, amount, time.Now())
}

// MarkUnpaid marks an attendee as unpaid
func (s *PaymentService) MarkUnpaid(ctx context.Context, attendeeID int64) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get attendee
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Mark as unpaid
		attendee.MarkUnpaid()

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

// WaivePayment waives payment for an attendee
func (s *PaymentService) WaivePayment(ctx context.Context, attendeeID int64, waiveDate time.Time) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get attendee
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Waive payment
		attendee.WaivePayment(waiveDate)

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

// WaivePaymentNow waives payment with the current timestamp
func (s *PaymentService) WaivePaymentNow(ctx context.Context, attendeeID int64) error {
	return s.WaivePayment(ctx, attendeeID, time.Now())
}

// UpdatePaymentAmount updates the payment amount for an attendee
func (s *PaymentService) UpdatePaymentAmount(ctx context.Context, attendeeID int64, amount decimal.Decimal) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get attendee
		attendee, err := s.attendeeRepo.GetByID(txCtx, attendeeID)
		if err != nil {
			return fmt.Errorf("attendee not found: %w", err)
		}
		if attendee == nil {
			return util.ErrAttendeeNotFound
		}

		// Update amount
		if err := attendee.SetPaymentAmount(amount); err != nil {
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

// GetPaymentSummary retrieves payment summary for an activity
func (s *PaymentService) GetPaymentSummary(ctx context.Context, activityID int64) (*repository.PaymentSummary, error) {
	// Validate activity exists
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}

	// Get payment summary
	summary, err := s.attendeeRepo.GetPaymentSummary(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}

	return summary, nil
}

// GetUnpaidAttendees retrieves all unpaid attendees for an activity
func (s *PaymentService) GetUnpaidAttendees(ctx context.Context, activityID int64) ([]*domain.Attendee, error) {
	// Validate activity exists
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}

	// Get unpaid attendees
	attendees, err := s.attendeeRepo.FindByPaymentStatus(ctx, activityID, domain.PaymentStatusUnpaid)
	if err != nil {
		return nil, fmt.Errorf("failed to get unpaid attendees: %w", err)
	}

	return attendees, nil
}

// GetPaidAttendees retrieves all paid attendees for an activity
func (s *PaymentService) GetPaidAttendees(ctx context.Context, activityID int64) ([]*domain.Attendee, error) {
	// Validate activity exists
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}

	// Get paid attendees
	attendees, err := s.attendeeRepo.FindByPaymentStatus(ctx, activityID, domain.PaymentStatusPaid)
	if err != nil {
		return nil, fmt.Errorf("failed to get paid attendees: %w", err)
	}

	return attendees, nil
}

// GetWaivedAttendees retrieves all attendees with waived payments for an activity
func (s *PaymentService) GetWaivedAttendees(ctx context.Context, activityID int64) ([]*domain.Attendee, error) {
	// Validate activity exists
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}

	// Get waived attendees
	attendees, err := s.attendeeRepo.FindByPaymentStatus(ctx, activityID, domain.PaymentStatusWaived)
	if err != nil {
		return nil, fmt.Errorf("failed to get waived attendees: %w", err)
	}

	return attendees, nil
}

// BulkMarkPaid marks multiple attendees as paid
func (s *PaymentService) BulkMarkPaid(ctx context.Context, attendeeIDs []int64, amount decimal.Decimal, paymentDate time.Time) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, attendeeID := range attendeeIDs {
			if err := s.MarkPaid(txCtx, attendeeID, amount, paymentDate); err != nil {
				return fmt.Errorf("failed to mark attendee %d as paid: %w", attendeeID, err)
			}
		}
		return nil
	})
}

// BulkWaivePayment waives payment for multiple attendees
func (s *PaymentService) BulkWaivePayment(ctx context.Context, attendeeIDs []int64, waiveDate time.Time) error {
	return s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, attendeeID := range attendeeIDs {
			if err := s.WaivePayment(txCtx, attendeeID, waiveDate); err != nil {
				return fmt.Errorf("failed to waive payment for attendee %d: %w", attendeeID, err)
			}
		}
		return nil
	})
}
