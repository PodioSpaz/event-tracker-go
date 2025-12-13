package domain

import (
	"github.com/PodioSpaz/event-tracker-go/internal/util"
)

// Role represents a configurable attendee role
type Role struct {
	ID          int64
	Name        string // Unique identifier (participant, volunteer, etc.)
	DisplayName string // Human-readable name
	Description string
	Active      bool
}

// NewRole creates a new Role
func NewRole(name, displayName, description string) *Role {
	return &Role{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Active:      true,
	}
}

// IsActive returns true if the role is active
func (r *Role) IsActive() bool {
	return r.Active
}

// Activate activates the role
func (r *Role) Activate() {
	r.Active = true
}

// Deactivate deactivates the role
func (r *Role) Deactivate() {
	r.Active = false
}

// ToAttendeeRole converts this Role to an AttendeeRole type
func (r *Role) ToAttendeeRole() AttendeeRole {
	return AttendeeRole(r.Name)
}

// Validate validates the role data
func (r *Role) Validate() error {
	// Required fields
	if err := util.ValidateRequired(r.Name, "name"); err != nil {
		return err
	}

	if err := util.ValidateRequired(r.DisplayName, "display_name"); err != nil {
		return err
	}

	// String length constraints
	if err := util.ValidateStringLength(r.Name, "name", 1, 50); err != nil {
		return err
	}

	if err := util.ValidateStringLength(r.DisplayName, "display_name", 1, 100); err != nil {
		return err
	}

	if r.Description != "" {
		if err := util.ValidateStringLength(r.Description, "description", 0, 500); err != nil {
			return err
		}
	}

	return nil
}

// DefaultRoles returns the default roles for the application
func DefaultRoles() []*Role {
	return []*Role{
		{
			Name:        "participant",
			DisplayName: "Participant",
			Description: "Regular event participant",
			Active:      true,
		},
		{
			Name:        "volunteer",
			DisplayName: "Volunteer",
			Description: "Event volunteer helper",
			Active:      true,
		},
		{
			Name:        "worship_team",
			DisplayName: "Worship Team",
			Description: "Worship team member",
			Active:      true,
		},
		{
			Name:        "workshop_leader",
			DisplayName: "Workshop Leader",
			Description: "Leads workshops during the event",
			Active:      true,
		},
	}
}
