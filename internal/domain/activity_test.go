package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewActivity(t *testing.T) {
	name := "Test Event"
	location := "Test Location"
	date := time.Now().AddDate(0, 0, 7)

	activity := NewActivity(name, location, date)

	assert.NotNil(t, activity)
	assert.Equal(t, name, activity.Name)
	assert.Equal(t, location, activity.Location)
	assert.Equal(t, date, activity.Date)
	assert.Equal(t, ActivityTypeEvent, activity.ActivityType)
	assert.Equal(t, StatusActive, activity.Status)
	assert.True(t, activity.RequiresRegistration)
	assert.False(t, activity.IsFree)
	assert.True(t, activity.Fee.IsZero())
}

func TestActivity_IsEvent(t *testing.T) {
	activity := &Activity{ActivityType: ActivityTypeEvent}
	assert.True(t, activity.IsEvent())

	activity.ActivityType = ActivityTypeGathering
	assert.False(t, activity.IsEvent())
}

func TestActivity_IsGathering(t *testing.T) {
	activity := &Activity{ActivityType: ActivityTypeGathering}
	assert.True(t, activity.IsGathering())

	activity.ActivityType = ActivityTypeEvent
	assert.False(t, activity.IsGathering())
}

func TestActivity_CanRegisterAttendees(t *testing.T) {
	tests := []struct {
		name                 string
		activityType         ActivityType
		requiresRegistration bool
		status               ActivityStatus
		expected             bool
	}{
		{"Event with registration active", ActivityTypeEvent, true, StatusActive, true},
		{"Event without registration", ActivityTypeEvent, false, StatusActive, false},
		{"Event cancelled", ActivityTypeEvent, true, StatusCancelled, false},
		{"Event completed", ActivityTypeEvent, true, StatusCompleted, false},
		{"Gathering", ActivityTypeGathering, true, StatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := &Activity{
				ActivityType:         tt.activityType,
				RequiresRegistration: tt.requiresRegistration,
				Status:               tt.status,
			}
			assert.Equal(t, tt.expected, activity.CanRegisterAttendees())
		})
	}
}

func TestActivity_IsFull(t *testing.T) {
	capacity := 10

	tests := []struct {
		name             string
		maxCapacity      *int
		currentAttendees int
		expected         bool
	}{
		{"Below capacity", &capacity, 5, false},
		{"At capacity", &capacity, 10, true},
		{"Over capacity", &capacity, 15, true},
		{"Unlimited capacity", nil, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := &Activity{
				ActivityType: ActivityTypeEvent,
				MaxCapacity:  tt.maxCapacity,
			}
			assert.Equal(t, tt.expected, activity.IsFull(tt.currentAttendees))
		})
	}
}

func TestActivity_GetCapacityRemaining(t *testing.T) {
	capacity := 10

	tests := []struct {
		name             string
		maxCapacity      *int
		currentAttendees int
		expected         int
	}{
		{"Some remaining", &capacity, 5, 5},
		{"None remaining", &capacity, 10, 0},
		{"Over capacity", &capacity, 15, 0},
		{"Unlimited", nil, 100, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := &Activity{MaxCapacity: tt.maxCapacity}
			assert.Equal(t, tt.expected, activity.GetCapacityRemaining(tt.currentAttendees))
		})
	}
}

func TestActivity_HasCapacityFor(t *testing.T) {
	capacity := 10

	tests := []struct {
		name                 string
		maxCapacity          *int
		currentAttendees     int
		additionalAttendees  int
		expected             bool
	}{
		{"Can fit", &capacity, 5, 3, true},
		{"Exactly fits", &capacity, 5, 5, true},
		{"Cannot fit", &capacity, 8, 5, false},
		{"Unlimited", nil, 100, 50, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := &Activity{MaxCapacity: tt.maxCapacity}
			result := activity.HasCapacityFor(tt.currentAttendees, tt.additionalAttendees)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivity_Cancel(t *testing.T) {
	activity := &Activity{Status: StatusActive}

	err := activity.Cancel()
	assert.NoError(t, err)
	assert.Equal(t, StatusCancelled, activity.Status)

	// Cannot cancel completed activity
	activity.Status = StatusCompleted
	err = activity.Cancel()
	assert.Error(t, err)
}

func TestActivity_Complete(t *testing.T) {
	activity := &Activity{Status: StatusActive}

	err := activity.Complete()
	assert.NoError(t, err)
	assert.Equal(t, StatusCompleted, activity.Status)

	// Cannot complete cancelled activity
	activity.Status = StatusCancelled
	err = activity.Complete()
	assert.Error(t, err)
}

func TestActivity_Reactivate(t *testing.T) {
	activity := &Activity{Status: StatusCancelled}

	err := activity.Reactivate()
	assert.NoError(t, err)
	assert.Equal(t, StatusActive, activity.Status)

	// Cannot reactivate completed activity
	activity.Status = StatusCompleted
	err = activity.Reactivate()
	assert.Error(t, err)
}

func TestActivity_SetFee(t *testing.T) {
	activity := &Activity{}

	// Valid fee
	err := activity.SetFee(decimal.NewFromFloat(25.50))
	assert.NoError(t, err)
	assert.Equal(t, "25.5", activity.Fee.String())
	assert.False(t, activity.IsFree)

	// Zero fee
	err = activity.SetFee(decimal.Zero)
	assert.NoError(t, err)
	assert.True(t, activity.IsFree)

	// Negative fee
	err = activity.SetFee(decimal.NewFromFloat(-10))
	assert.Error(t, err)
}

func TestActivity_SetCapacity(t *testing.T) {
	activity := &Activity{}

	// Valid capacity
	capacity := 50
	err := activity.SetCapacity(&capacity)
	assert.NoError(t, err)
	assert.NotNil(t, activity.MaxCapacity)
	assert.Equal(t, 50, *activity.MaxCapacity)

	// Unlimited capacity
	err = activity.SetCapacity(nil)
	assert.NoError(t, err)
	assert.Nil(t, activity.MaxCapacity)

	// Negative capacity
	negCapacity := -10
	err = activity.SetCapacity(&negCapacity)
	assert.Error(t, err)
}

func TestActivity_Validate(t *testing.T) {
	validActivity := &Activity{
		Name:         "Test Event",
		Location:     "Test Location",
		Date:         time.Now().AddDate(0, 0, 7),
		ActivityType: ActivityTypeEvent,
		Status:       StatusActive,
		Fee:          decimal.Zero,
	}

	err := validActivity.Validate()
	assert.NoError(t, err)

	// Missing name
	invalidActivity := *validActivity
	invalidActivity.Name = ""
	assert.Error(t, invalidActivity.Validate())

	// Missing location
	invalidActivity = *validActivity
	invalidActivity.Location = ""
	assert.Error(t, invalidActivity.Validate())

	// Missing date
	invalidActivity = *validActivity
	invalidActivity.Date = time.Time{}
	assert.Error(t, invalidActivity.Validate())

	// Invalid activity type
	invalidActivity = *validActivity
	invalidActivity.ActivityType = "invalid"
	assert.Error(t, invalidActivity.Validate())

	// Invalid status
	invalidActivity = *validActivity
	invalidActivity.Status = "invalid"
	assert.Error(t, invalidActivity.Validate())

	// Negative fee
	invalidActivity = *validActivity
	invalidActivity.Fee = decimal.NewFromFloat(-10)
	assert.Error(t, invalidActivity.Validate())
}

func TestActivity_ValidateForRegistration(t *testing.T) {
	validActivity := &Activity{
		ActivityType:         ActivityTypeEvent,
		RequiresRegistration: true,
		Status:               StatusActive,
	}

	err := validActivity.ValidateForRegistration(0)
	assert.NoError(t, err)

	// Gathering
	activity := *validActivity
	activity.ActivityType = ActivityTypeGathering
	assert.Error(t, activity.ValidateForRegistration(0))

	// No registration required
	activity = *validActivity
	activity.RequiresRegistration = false
	assert.Error(t, activity.ValidateForRegistration(0))

	// Cancelled
	activity = *validActivity
	activity.Status = StatusCancelled
	assert.Error(t, activity.ValidateForRegistration(0))

	// Completed
	activity = *validActivity
	activity.Status = StatusCompleted
	assert.Error(t, activity.ValidateForRegistration(0))

	// Full capacity
	activity = *validActivity
	capacity := 10
	activity.MaxCapacity = &capacity
	assert.Error(t, activity.ValidateForRegistration(10))
}
