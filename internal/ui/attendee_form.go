package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
)

// AttendeeForm represents a form for registering attendees to activities
type AttendeeForm struct {
	app        *App
	activityID int64
	onSave     func()

	// Form fields
	personSelect  *widget.Select
	roleSelect    *widget.Select
	paymentSelect *widget.Select

	// Available people (for person selection)
	people     []*domain.Person
	peopleMap  map[string]*domain.Person
	peopleList []string

	// Form container
	formContainer *fyne.Container
}

// NewAttendeeForm creates a new attendee registration form
func NewAttendeeForm(app *App, activityID int64, onSave func()) *AttendeeForm {
	form := &AttendeeForm{
		app:        app,
		activityID: activityID,
		onSave:     onSave,
		peopleMap:  make(map[string]*domain.Person),
	}

	form.loadPeople()
	form.initFields()
	form.buildForm()

	return form
}

// loadPeople loads all available people from the database
func (f *AttendeeForm) loadPeople() {
	people, err := f.app.personCtrl.GetAllPeople(f.app.ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load people")
		return
	}

	f.people = people
	f.peopleList = make([]string, len(people))

	for i, person := range people {
		displayName := fmt.Sprintf("%s (%s)", person.FullName(), person.Email)
		f.peopleList[i] = displayName
		f.peopleMap[displayName] = person
	}
}

// initFields initializes all form fields
func (f *AttendeeForm) initFields() {
	// Person selection
	f.personSelect = widget.NewSelect(f.peopleList, nil)
	if len(f.peopleList) > 0 {
		f.personSelect.SetSelected(f.peopleList[0])
	}

	// Role selection
	f.roleSelect = widget.NewSelect(
		[]string{
			domain.RoleParticipant.DisplayName(),
			domain.RoleVolunteer.DisplayName(),
			domain.RoleWorshipTeam.DisplayName(),
			domain.RoleWorkshopLeader.DisplayName(),
		},
		nil,
	)
	f.roleSelect.SetSelected(domain.RoleParticipant.DisplayName())

	// Payment status selection
	f.paymentSelect = widget.NewSelect(
		[]string{"paid", "unpaid", "waived"},
		nil,
	)
	f.paymentSelect.SetSelected("unpaid")
}

// buildForm builds the form UI
func (f *AttendeeForm) buildForm() {
	// Add new person button
	addPersonBtn := widget.NewButton("+ Add New Person", func() {
		f.showAddPersonDialog()
	})
	addPersonBtn.Importance = widget.LowImportance

	f.formContainer = container.NewVBox(
		// Person selection
		widget.NewLabel("Person *"),
		f.personSelect,
		addPersonBtn,

		widget.NewSeparator(),

		// Role selection
		widget.NewLabel("Role *"),
		f.roleSelect,

		// Payment status
		widget.NewLabel("Payment Status *"),
		f.paymentSelect,
	)
}

// Show renders the form in a dialog
func (f *AttendeeForm) Show(parentWindow fyne.Window) {
	if len(f.peopleList) == 0 {
		dialog.ShowInformation(
			"No People Available",
			"Please add people to the database before registering attendees.",
			parentWindow,
		)
		return
	}

	// Create scrollable form
	scrollForm := container.NewVScroll(f.formContainer)
	scrollForm.SetMinSize(fyne.NewSize(400, 300))

	// Create dialog with save and cancel buttons
	var formDialog dialog.Dialog
	formDialog = dialog.NewCustomConfirm(
		"Register Attendee",
		"Register",
		"Cancel",
		scrollForm,
		func(save bool) {
			if save {
				f.handleSave()
			}
		},
		parentWindow,
	)

	formDialog.Resize(fyne.NewSize(450, 400))
	formDialog.Show()
}

// showAddPersonDialog shows a dialog to add a new person
func (f *AttendeeForm) showAddPersonDialog() {
	personForm := NewPersonForm(f.app, nil, func() {
		// Reload people list after adding a new person
		f.loadPeople()
		f.personSelect.Options = f.peopleList
		if len(f.peopleList) > 0 {
			f.personSelect.SetSelected(f.peopleList[len(f.peopleList)-1])
		}
		f.personSelect.Refresh()
	})
	personForm.Show(f.app.mainWindow)
}

// handleSave validates and saves the attendee registration
func (f *AttendeeForm) handleSave() {
	// Validate person selection
	if f.personSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("please select a person"), f.app.mainWindow)
		return
	}

	// Get selected person
	person, ok := f.peopleMap[f.personSelect.Selected]
	if !ok {
		dialog.ShowError(fmt.Errorf("invalid person selection"), f.app.mainWindow)
		return
	}

	// Map role display name to role value
	var role domain.AttendeeRole
	switch f.roleSelect.Selected {
	case domain.RoleParticipant.DisplayName():
		role = domain.RoleParticipant
	case domain.RoleVolunteer.DisplayName():
		role = domain.RoleVolunteer
	case domain.RoleWorshipTeam.DisplayName():
		role = domain.RoleWorshipTeam
	case domain.RoleWorkshopLeader.DisplayName():
		role = domain.RoleWorkshopLeader
	default:
		role = domain.RoleParticipant
	}

	// Register attendee using the controller
	attendee, err := f.app.activityCtrl.RegisterAttendee(f.app.ctx, f.activityID, person.ID, role)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register attendee")
		dialog.ShowError(fmt.Errorf("failed to register attendee: %w", err), f.app.mainWindow)
		return
	}

	// Get activity to determine the fee amount
	activity, err := f.app.activityCtrl.GetActivity(f.app.ctx, f.activityID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get activity for payment processing")
		// Don't fail the registration, just log the error
	} else {
		// Update payment status based on selection
		paymentStatus := f.paymentSelect.Selected
		switch paymentStatus {
		case "paid":
			// Mark as paid with the activity fee
			if err := f.app.activityCtrl.MarkAttendeePaid(f.app.ctx, attendee.ID, activity.Fee); err != nil {
				log.Error().Err(err).Msg("Failed to mark attendee as paid")
				dialog.ShowError(fmt.Errorf("attendee registered but failed to mark as paid: %w", err), f.app.mainWindow)
				return
			}
		case "waived":
			// Waive payment
			if err := f.app.activityCtrl.WaiveAttendeePayment(f.app.ctx, attendee.ID); err != nil {
				log.Error().Err(err).Msg("Failed to waive attendee payment")
				dialog.ShowError(fmt.Errorf("attendee registered but failed to waive payment: %w", err), f.app.mainWindow)
				return
			}
		case "unpaid":
			// Default status, nothing to do
		}
	}

	log.Info().
		Str("person", person.FullName()).
		Int64("activity_id", f.activityID).
		Str("role", string(role)).
		Msg("Attendee registered successfully")

	dialog.ShowInformation("Success", "Attendee registered successfully", f.app.mainWindow)

	// Call onSave callback to refresh the view
	if f.onSave != nil {
		f.onSave()
	}
}
