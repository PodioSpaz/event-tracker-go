package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createNavigation creates the navigation sidebar
func (a *App) createNavigation() fyne.CanvasObject {
	// Create navigation buttons
	dashboardBtn := widget.NewButton("Dashboard", func() {
		a.showDashboard()
	})
	dashboardBtn.Importance = widget.HighImportance

	activitiesBtn := widget.NewButton("Activities", func() {
		a.showActivities()
	})

	peopleBtn := widget.NewButton("People", func() {
		a.showPeople()
	})

	// Navigation container
	nav := container.NewVBox(
		widget.NewLabelWithStyle("Event Tracker", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		dashboardBtn,
		activitiesBtn,
		peopleBtn,
		layout.NewSpacer(),
		widget.NewSeparator(),
		widget.NewLabel("v0.1.0-dev"),
	)

	return nav
}

// Navigation methods to show different views

func (a *App) showDashboard() {
	view := NewDashboardView(a)
	a.setContent(view.Render(), "dashboard")
}

func (a *App) showActivities() {
	view := NewActivitiesView(a)
	a.setContent(view.Render(), "activities")
}

func (a *App) showPeople() {
	view := NewPeopleView(a)
	a.setContent(view.Render(), "people")
}

func (a *App) showActivityDetail(activityID int64) {
	view := NewActivityDetailView(a, activityID)
	a.setContent(view.Render(), "activity-detail")
}

func (a *App) showPersonDetail(personID int64) {
	view := NewPersonDetailView(a, personID)
	a.setContent(view.Render(), "person-detail")
}
