package service

import (
	"context"
	"fmt"

	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/repository"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// CapacityInfo contains capacity information for an activity
type CapacityInfo struct {
	MaxCapacity       *int // nil for unlimited
	CurrentAttendees  int
	Remaining         int  // -1 for unlimited
	IsUnlimited       bool
	IsFull            bool
	PercentageFilled  float64
}

// CapacityService handles capacity checking and management
type CapacityService struct {
	activityRepo repository.ActivityRepository
	attendeeRepo repository.AttendeeRepository
}

// NewCapacityService creates a new capacity service
func NewCapacityService(
	activityRepo repository.ActivityRepository,
	attendeeRepo repository.AttendeeRepository,
) *CapacityService {
	return &CapacityService{
		activityRepo: activityRepo,
		attendeeRepo: attendeeRepo,
	}
}

// GetCapacityInfo retrieves capacity information for an activity
func (s *CapacityService) GetCapacityInfo(ctx context.Context, activityID int64) (*CapacityInfo, error) {
	// Get activity
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return nil, util.ErrActivityNotFound
	}

	// Get current attendee count
	currentCount, err := s.attendeeRepo.CountByActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendee count: %w", err)
	}

	// Build capacity info
	info := &CapacityInfo{
		MaxCapacity:      activity.MaxCapacity,
		CurrentAttendees: currentCount,
		IsUnlimited:      activity.MaxCapacity == nil,
		IsFull:           activity.IsFull(currentCount),
	}

	// Calculate remaining capacity
	if activity.MaxCapacity == nil {
		info.Remaining = -1 // Unlimited
		info.PercentageFilled = 0.0
	} else {
		info.Remaining = activity.GetCapacityRemaining(currentCount)
		if *activity.MaxCapacity > 0 {
			info.PercentageFilled = (float64(currentCount) / float64(*activity.MaxCapacity)) * 100.0
		}
	}

	return info, nil
}

// CanAcceptAttendees checks if an activity can accept N more attendees
func (s *CapacityService) CanAcceptAttendees(ctx context.Context, activityID int64, count int) (bool, error) {
	// Get activity
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return false, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return false, util.ErrActivityNotFound
	}

	// Get current attendee count
	currentCount, err := s.attendeeRepo.CountByActivity(ctx, activityID)
	if err != nil {
		return false, fmt.Errorf("failed to get attendee count: %w", err)
	}

	// Check capacity
	return activity.HasCapacityFor(currentCount, count), nil
}

// IsFull checks if an activity is at capacity
func (s *CapacityService) IsFull(ctx context.Context, activityID int64) (bool, error) {
	// Get activity
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return false, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return false, util.ErrActivityNotFound
	}

	// Get current attendee count
	currentCount, err := s.attendeeRepo.CountByActivity(ctx, activityID)
	if err != nil {
		return false, fmt.Errorf("failed to get attendee count: %w", err)
	}

	return activity.IsFull(currentCount), nil
}

// GetRemainingCapacity returns the number of available spots
func (s *CapacityService) GetRemainingCapacity(ctx context.Context, activityID int64) (int, error) {
	// Get activity
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return 0, fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return 0, util.ErrActivityNotFound
	}

	// Get current attendee count
	currentCount, err := s.attendeeRepo.CountByActivity(ctx, activityID)
	if err != nil {
		return 0, fmt.Errorf("failed to get attendee count: %w", err)
	}

	return activity.GetCapacityRemaining(currentCount), nil
}

// ValidateCapacityForRegistration validates that an activity can accept registrations
func (s *CapacityService) ValidateCapacityForRegistration(ctx context.Context, activityID int64) error {
	// Get activity
	activity, err := s.activityRepo.GetByID(ctx, activityID)
	if err != nil {
		return fmt.Errorf("activity not found: %w", err)
	}
	if activity == nil {
		return util.ErrActivityNotFound
	}

	// Get current attendee count
	currentCount, err := s.attendeeRepo.CountByActivity(ctx, activityID)
	if err != nil {
		return fmt.Errorf("failed to get attendee count: %w", err)
	}

	// Validate activity can accept registrations
	return activity.ValidateForRegistration(currentCount)
}

// GetCapacityStats returns capacity statistics for multiple activities
func (s *CapacityService) GetCapacityStats(ctx context.Context, activityIDs []int64) (map[int64]*CapacityInfo, error) {
	stats := make(map[int64]*CapacityInfo)

	for _, activityID := range activityIDs {
		info, err := s.GetCapacityInfo(ctx, activityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get capacity info for activity %d: %w", activityID, err)
		}
		stats[activityID] = info
	}

	return stats, nil
}

// GetNearCapacityActivities returns activities that are near capacity (above threshold percentage)
func (s *CapacityService) GetNearCapacityActivities(ctx context.Context, thresholdPercent float64) ([]*domain.Activity, error) {
	// Get all active activities
	activities, err := s.activityRepo.FindByStatus(ctx, domain.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	var nearCapacity []*domain.Activity

	for _, activity := range activities {
		// Skip activities with unlimited capacity
		if activity.MaxCapacity == nil {
			continue
		}

		// Get capacity info
		info, err := s.GetCapacityInfo(ctx, activity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get capacity info: %w", err)
		}

		// Check if near capacity
		if info.PercentageFilled >= thresholdPercent {
			nearCapacity = append(nearCapacity, activity)
		}
	}

	return nearCapacity, nil
}

// GetFullActivities returns all activities that are at full capacity
func (s *CapacityService) GetFullActivities(ctx context.Context) ([]*domain.Activity, error) {
	// Get all active activities
	activities, err := s.activityRepo.FindByStatus(ctx, domain.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	var fullActivities []*domain.Activity

	for _, activity := range activities {
		// Get current attendee count
		currentCount, err := s.attendeeRepo.CountByActivity(ctx, activity.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get attendee count: %w", err)
		}

		// Check if full
		if activity.IsFull(currentCount) {
			fullActivities = append(fullActivities, activity)
		}
	}

	return fullActivities, nil
}
