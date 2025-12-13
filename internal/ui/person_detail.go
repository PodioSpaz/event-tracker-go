package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
)

// PersonDetailView displays details of a single person
type PersonDetailView struct {
	app      *App
	personID int64
}

// NewPersonDetailView creates a new person detail view
func NewPersonDetailView(app *App, personID int64) *PersonDetailView {
	return &PersonDetailView{
		app:      app,
		personID: personID,
	}
}

// Render renders the person detail view
func (v *PersonDetailView) Render() fyne.CanvasObject {
	// Get person
	person, err := v.app.personCtrl.GetPerson(v.app.ctx, v.personID)
	if err != nil {
		log.Error().Err(err).Int64("id", v.personID).Msg("Failed to load person")
		return widget.NewLabel("Error loading person")
	}

	// Back button
	backBtn := widget.NewButton("← Back to People", func() {
		v.app.showPeople()
	})

	// Edit button
	editBtn := widget.NewButton("Edit", func() {
		v.showEditPersonDialog(person)
	})
	editBtn.Importance = widget.HighImportance

	// Header
	header := widget.NewLabelWithStyle(person.FullName(), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Header with buttons
	headerContainer := container.NewBorder(nil, nil, backBtn, editBtn, header)

	// Details
	details := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Email: %s", person.Email)),
		widget.NewLabel(fmt.Sprintf("Phone: %s", person.Phone)),
	)

	// Registrations section
	registrationsHeader := widget.NewLabelWithStyle("Activity Registrations", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	registrationsList := v.createRegistrationsList()

	// Main content
	content := container.NewBorder(
		container.NewVBox(headerContainer, widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewVScroll(
			container.NewVBox(
				details,
				widget.NewSeparator(),
				registrationsHeader,
				registrationsList,
			),
		),
	)

	return content
}

// showEditPersonDialog shows dialog to edit person
func (v *PersonDetailView) showEditPersonDialog(person *domain.Person) {
	form := NewPersonForm(v.app, person, func() {
		// Refresh the person detail view after saving
		v.app.showPersonDetail(person.ID)
	})
	form.Show(v.app.mainWindow)
}

// createRegistrationsList creates a list of activity registrations for this person
func (v *PersonDetailView) createRegistrationsList() fyne.CanvasObject {
	registrations, err := v.app.personCtrl.GetPersonRegistrations(v.app.ctx, v.personID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load registrations")
		return widget.NewLabel("Error loading registrations")
	}

	if len(registrations) == 0 {
		return widget.NewLabel("No activity registrations")
	}

	items := make([]fyne.CanvasObject, 0, len(registrations))
	for _, reg := range registrations {
		// Get activity details to show activity name
		activity, err := v.app.activityCtrl.GetActivity(v.app.ctx, reg.ActivityID)
		if err != nil {
			log.Error().Err(err).Int64("activity_id", reg.ActivityID).Msg("Failed to load activity")
			continue
		}

		roleText := reg.GetRoleDisplay()
		statusText := string(reg.PaymentStatus)

		item := widget.NewLabel(fmt.Sprintf("• %s - %s (%s) - %s",
			activity.Name,
			roleText,
			statusText,
			activity.Date.Format("Jan 2, 2006")))
		items = append(items, item)
	}

	return container.NewVBox(items...)
}
