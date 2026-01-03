package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
)

// ActivitiesView displays the list of activities
type ActivitiesView struct {
	app *App
}

// NewActivitiesView creates a new activities view
func NewActivitiesView(app *App) *ActivitiesView {
	return &ActivitiesView{app: app}
}

// Render renders the activities list view
func (v *ActivitiesView) Render() fyne.CanvasObject {
	// Header with add button
	header := widget.NewLabelWithStyle("Activities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addBtn := widget.NewButton("New Activity", func() {
		v.showNewActivityDialog()
	})
	addBtn.Importance = widget.HighImportance

	headerContainer := container.NewBorder(nil, nil, header, addBtn)

	// Get activities
	activities, err := v.app.activityCtrl.GetAllActivities(v.app.ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load activities")
		return widget.NewLabel("Error loading activities")
	}

	// Create table
	table := v.createActivitiesTable(activities)

	// Main content
	content := container.NewBorder(
		container.NewVBox(headerContainer, widget.NewSeparator()),
		nil,
		nil,
		nil,
		table,
	)

	return content
}

// createActivitiesTable creates a table of activities
func (v *ActivitiesView) createActivitiesTable(activities []*domain.Activity) fyne.CanvasObject {
	if len(activities) == 0 {
		return widget.NewLabel("No activities found. Create your first activity!")
	}

	// Create table data
	data := make([][]string, len(activities)+1)

	// Header row
	data[0] = []string{"Name", "Date", "Location", "Type", "Status", "Capacity", "Actions"}

	// Data rows
	for i, activity := range activities {
		capacityStr := "Unlimited"
		if activity.MaxCapacity != nil {
			capacityStr = fmt.Sprintf("%d", *activity.MaxCapacity)
		}

		data[i+1] = []string{
			activity.Name,
			activity.Date.Format("Jan 2, 2006"),
			activity.Location,
			string(activity.ActivityType),
			string(activity.Status),
			capacityStr,
			"View",
		}
	}

	// Create table widget
	table := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			// Header row styling
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(data[id.Row][id.Col])
				return
			}

			// Action column
			if id.Col == 6 {
				// We'll show just text for now, buttons in tables are complex in Fyne
				label.SetText("→")
				return
			}

			label.SetText(data[id.Row][id.Col])
		},
	)

	// Refresh table to ensure initial render with data
	table.Refresh()

	// Set column widths
	table.SetColumnWidth(0, 200) // Name
	table.SetColumnWidth(1, 120) // Date
	table.SetColumnWidth(2, 150) // Location
	table.SetColumnWidth(3, 100) // Type
	table.SetColumnWidth(4, 100) // Status
	table.SetColumnWidth(5, 100) // Capacity
	table.SetColumnWidth(6, 80)  // Actions

	// Handle row selection
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 && id.Row <= len(activities) {
			activity := activities[id.Row-1]
			v.app.showActivityDetail(activity.ID)
		}
	}

	return container.NewScroll(table)
}

// showNewActivityDialog shows dialog to create new activity
func (v *ActivitiesView) showNewActivityDialog() {
	form := NewActivityForm(v.app, nil, func() {
		// Refresh the activities view after saving
		v.app.showActivities()
	})
	form.Show(v.app.mainWindow)
}
