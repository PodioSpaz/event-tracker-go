package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
)

// DashboardView displays the main dashboard
type DashboardView struct {
	app *App
}

// NewDashboardView creates a new dashboard view
func NewDashboardView(app *App) *DashboardView {
	return &DashboardView{app: app}
}

// Render renders the dashboard view
func (v *DashboardView) Render() fyne.CanvasObject {
	// Header
	header := widget.NewLabelWithStyle("Dashboard", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Get statistics
	stats := v.getStatistics()

	// Create stat cards
	statsContainer := container.NewGridWithColumns(4,
		v.createStatCard("Activities", fmt.Sprintf("%d", stats.TotalActivities), theme.ColorNamePrimary),
		v.createStatCard("People", fmt.Sprintf("%d", stats.TotalPeople), theme.ColorNameSuccess),
		v.createStatCard("Attendees", fmt.Sprintf("%d", stats.TotalAttendees), theme.ColorNameWarning),
		v.createStatCard("Upcoming", fmt.Sprintf("%d", stats.UpcomingActivities), theme.ColorNameError),
	)

	// Upcoming activities section
	upcomingHeader := widget.NewLabelWithStyle("Upcoming Activities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	upcomingList := v.createUpcomingActivitiesList()

	// Recent activity section
	recentHeader := widget.NewLabelWithStyle("Recent Activities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	recentList := v.createRecentActivitiesList()

	// Create two-column layout for lists
	listsContainer := container.NewGridWithColumns(2,
		container.NewBorder(upcomingHeader, nil, nil, nil, upcomingList),
		container.NewBorder(recentHeader, nil, nil, nil, recentList),
	)

	// Main content
	content := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator(), statsContainer, widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewVScroll(listsContainer),
	)

	return content
}

// createStatCard creates a statistics card
func (v *DashboardView) createStatCard(label, value string, colorName fyne.ThemeColorName) fyne.CanvasObject {
	valueLabel := widget.NewLabelWithStyle(value, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	nameLabel := widget.NewLabel(label)
	nameLabel.Alignment = fyne.TextAlignCenter

	// Create colored background
	bg := canvas.NewRectangle(theme.Color(colorName))
	bg.SetMinSize(fyne.NewSize(200, 100))

	card := container.NewMax(
		bg,
		container.NewVBox(
			layout.NewSpacer(),
			valueLabel,
			nameLabel,
			layout.NewSpacer(),
		),
	)

	return card
}

// Statistics holds dashboard statistics
type Statistics struct {
	TotalActivities    int
	TotalPeople        int
	TotalAttendees     int
	UpcomingActivities int
	ActiveActivities   int
}

// getStatistics retrieves dashboard statistics
func (v *DashboardView) getStatistics() Statistics {
	stats := Statistics{}

	// Get total activities
	count, err := v.app.activityCtrl.GetAllActivities(v.app.ctx)
	if err == nil {
		stats.TotalActivities = len(count)
	}

	// Get total people
	people, err := v.app.personCtrl.GetAllPeople(v.app.ctx)
	if err == nil {
		stats.TotalPeople = len(people)
	}

	// Get upcoming activities
	upcoming, err := v.app.activityCtrl.GetUpcomingActivities(v.app.ctx, 10)
	if err == nil {
		stats.UpcomingActivities = len(upcoming)
	}

	// Get active activities
	active, err := v.app.activityCtrl.GetActivitiesByStatus(v.app.ctx, domain.StatusActive)
	if err == nil {
		stats.ActiveActivities = len(active)
	}

	// Placeholder for attendees count
	stats.TotalAttendees = 0

	return stats
}

// createUpcomingActivitiesList creates a list of upcoming activities
func (v *DashboardView) createUpcomingActivitiesList() fyne.CanvasObject {
	activities, err := v.app.activityCtrl.GetUpcomingActivities(v.app.ctx, 5)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get upcoming activities")
		return widget.NewLabel("Error loading activities")
	}

	if len(activities) == 0 {
		return widget.NewLabel("No upcoming activities")
	}

	items := make([]fyne.CanvasObject, 0, len(activities))
	for _, activity := range activities {
		activityCopy := activity // Capture for closure

		// Format date
		dateStr := activity.Date.Format("Jan 2, 2006")
		daysUntil := activity.DaysUntil()
		daysStr := ""
		if daysUntil == 0 {
			daysStr = "Today"
		} else if daysUntil == 1 {
			daysStr = "Tomorrow"
		} else if daysUntil > 0 {
			daysStr = fmt.Sprintf("In %d days", daysUntil)
		}

		nameLabel := widget.NewLabelWithStyle(activity.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		dateLabel := widget.NewLabel(fmt.Sprintf("%s - %s", dateStr, daysStr))
		locationLabel := widget.NewLabel(activity.Location)

		// Create clickable card
		card := widget.NewButton("", func() {
			v.app.showActivityDetail(activityCopy.ID)
		})
		card.Importance = widget.LowImportance

		item := container.NewBorder(
			container.NewVBox(nameLabel, dateLabel, locationLabel),
			nil, nil, nil,
			card,
		)

		items = append(items, item)
	}

	return container.NewVBox(items...)
}

// createRecentActivitiesList creates a list of recent activities
func (v *DashboardView) createRecentActivitiesList() fyne.CanvasObject {
	// Get recent completed or cancelled activities
	now := time.Now()
	past := now.AddDate(0, -3, 0) // Last 3 months

	activities, err := v.app.activityCtrl.GetActivitiesByDateRange(v.app.ctx, past, now)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent activities")
		return widget.NewLabel("Error loading activities")
	}

	// Filter to only completed/cancelled and limit to 5
	var recent []*domain.Activity
	for _, activity := range activities {
		if activity.IsCompleted() || activity.IsCancelled() {
			recent = append(recent, activity)
			if len(recent) >= 5 {
				break
			}
		}
	}

	if len(recent) == 0 {
		return widget.NewLabel("No recent activities")
	}

	items := make([]fyne.CanvasObject, 0, len(recent))
	for _, activity := range recent {
		activityCopy := activity // Capture for closure

		nameLabel := widget.NewLabelWithStyle(activity.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		dateLabel := widget.NewLabel(activity.Date.Format("Jan 2, 2006"))

		statusText := "Completed"
		if activity.IsCancelled() {
			statusText = "Cancelled"
		}
		statusLabel := widget.NewLabel(statusText)

		// Create clickable card
		card := widget.NewButton("", func() {
			v.app.showActivityDetail(activityCopy.ID)
		})
		card.Importance = widget.LowImportance

		item := container.NewBorder(
			container.NewVBox(nameLabel, dateLabel, statusLabel),
			nil, nil, nil,
			card,
		)

		items = append(items, item)
	}

	return container.NewVBox(items...)
}
