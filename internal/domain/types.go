package domain

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeEvent     ActivityType = "event"
	ActivityTypeGathering ActivityType = "gathering"
)

// String returns the string representation of ActivityType
func (t ActivityType) String() string {
	return string(t)
}

// IsValid checks if the activity type is valid
func (t ActivityType) IsValid() bool {
	switch t {
	case ActivityTypeEvent, ActivityTypeGathering:
		return true
	}
	return false
}

// ActivityStatus represents the current status of an activity
type ActivityStatus string

const (
	StatusActive    ActivityStatus = "active"
	StatusCancelled ActivityStatus = "cancelled"
	StatusCompleted ActivityStatus = "completed"
)

// String returns the string representation of ActivityStatus
func (s ActivityStatus) String() string {
	return string(s)
}

// IsValid checks if the activity status is valid
func (s ActivityStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusCancelled, StatusCompleted:
		return true
	}
	return false
}

// AttendeeRole represents the role of an attendee at an event
type AttendeeRole string

const (
	RoleParticipant    AttendeeRole = "participant"
	RoleVolunteer      AttendeeRole = "volunteer"
	RoleWorshipTeam    AttendeeRole = "worship_team"
	RoleWorkshopLeader AttendeeRole = "workshop_leader"
)

// String returns the string representation of AttendeeRole
func (r AttendeeRole) String() string {
	return string(r)
}

// DisplayName returns a human-readable display name for the role
func (r AttendeeRole) DisplayName() string {
	displays := map[AttendeeRole]string{
		RoleParticipant:    "Participant",
		RoleVolunteer:      "Volunteer",
		RoleWorshipTeam:    "Worship Team",
		RoleWorkshopLeader: "Workshop Leader",
	}
	if display, ok := displays[r]; ok {
		return display
	}
	return string(r)
}

// IsValid checks if the attendee role is valid
func (r AttendeeRole) IsValid() bool {
	switch r {
	case RoleParticipant, RoleVolunteer, RoleWorshipTeam, RoleWorkshopLeader:
		return true
	}
	return false
}

// PaymentStatus represents the payment status of an attendee
type PaymentStatus string

const (
	PaymentStatusPaid   PaymentStatus = "paid"
	PaymentStatusUnpaid PaymentStatus = "unpaid"
	PaymentStatusWaived PaymentStatus = "waived"
)

// String returns the string representation of PaymentStatus
func (p PaymentStatus) String() string {
	return string(p)
}

// IsValid checks if the payment status is valid
func (p PaymentStatus) IsValid() bool {
	switch p {
	case PaymentStatusPaid, PaymentStatusUnpaid, PaymentStatusWaived:
		return true
	}
	return false
}

// IsPaid returns true if the payment is considered paid (paid or waived)
func (p PaymentStatus) IsPaid() bool {
	return p == PaymentStatusPaid || p == PaymentStatusWaived
}

// ExpenditureCategory represents the category of an expenditure
type ExpenditureCategory string

const (
	CategoryVenue     ExpenditureCategory = "venue"
	CategoryFood      ExpenditureCategory = "food"
	CategorySupplies  ExpenditureCategory = "supplies"
	CategoryTransport ExpenditureCategory = "transport"
	CategoryOther     ExpenditureCategory = "other"
)

// String returns the string representation of ExpenditureCategory
func (c ExpenditureCategory) String() string {
	return string(c)
}

// IsValid checks if the expenditure category is valid
func (c ExpenditureCategory) IsValid() bool {
	switch c {
	case CategoryVenue, CategoryFood, CategorySupplies, CategoryTransport, CategoryOther:
		return true
	}
	return false
}
