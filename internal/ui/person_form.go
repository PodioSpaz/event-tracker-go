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

// PersonForm represents a form for creating or editing people
type PersonForm struct {
	app    *App
	person *domain.Person // nil for create, populated for edit
	onSave func()

	// Form fields
	firstNameEntry *widget.Entry
	lastNameEntry  *widget.Entry
	emailEntry     *widget.Entry
	phoneEntry     *widget.Entry

	// Form container
	formContainer *fyne.Container
}

// NewPersonForm creates a new person form
func NewPersonForm(app *App, person *domain.Person, onSave func()) *PersonForm {
	form := &PersonForm{
		app:    app,
		person: person,
		onSave: onSave,
	}

	form.initFields()
	form.buildForm()

	return form
}

// initFields initializes all form fields
func (f *PersonForm) initFields() {
	// First Name
	f.firstNameEntry = widget.NewEntry()
	f.firstNameEntry.SetPlaceHolder("Enter first name")

	// Last Name
	f.lastNameEntry = widget.NewEntry()
	f.lastNameEntry.SetPlaceHolder("Enter last name")

	// Email
	f.emailEntry = widget.NewEntry()
	f.emailEntry.SetPlaceHolder("email@example.com")

	// Phone
	f.phoneEntry = widget.NewEntry()
	f.phoneEntry.SetPlaceHolder("123-456-7890")

	// Populate fields if editing
	if f.person != nil {
		f.populateFields()
	}
}

// populateFields populates form fields from the person
func (f *PersonForm) populateFields() {
	f.firstNameEntry.SetText(f.person.FirstName)
	f.lastNameEntry.SetText(f.person.LastName)
	f.emailEntry.SetText(f.person.Email)
	f.phoneEntry.SetText(f.person.Phone)
}

// buildForm builds the form UI
func (f *PersonForm) buildForm() {
	f.formContainer = container.NewVBox(
		// First Name
		widget.NewLabel("First Name *"),
		f.firstNameEntry,

		// Last Name
		widget.NewLabel("Last Name *"),
		f.lastNameEntry,

		// Email
		widget.NewLabel("Email"),
		f.emailEntry,

		// Phone
		widget.NewLabel("Phone"),
		f.phoneEntry,

		widget.NewSeparator(),
		widget.NewLabel("* At least one contact method (email or phone) is required"),
	)
}

// Show renders the form in a dialog
func (f *PersonForm) Show(parentWindow fyne.Window) {
	title := "New Person"
	if f.person != nil {
		title = "Edit Person"
	}

	// Create scrollable form
	scrollForm := container.NewVScroll(f.formContainer)
	scrollForm.SetMinSize(fyne.NewSize(400, 400))

	// Create dialog with save and cancel buttons
	var formDialog dialog.Dialog
	formDialog = dialog.NewCustomConfirm(
		title,
		"Save",
		"Cancel",
		scrollForm,
		func(save bool) {
			if save {
				f.handleSave()
			}
		},
		parentWindow,
	)

	formDialog.Resize(fyne.NewSize(450, 450))
	formDialog.Show()
}

// handleSave validates and saves the person
func (f *PersonForm) handleSave() {
	// Validate required fields
	if f.firstNameEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("first name is required"), f.app.mainWindow)
		return
	}

	if f.lastNameEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("last name is required"), f.app.mainWindow)
		return
	}

	// Validate at least one contact method
	if f.emailEntry.Text == "" && f.phoneEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("at least one contact method (email or phone) is required"), f.app.mainWindow)
		return
	}

	// Create or update person
	if f.person == nil {
		// Create new person
		person := domain.NewPerson(
			f.firstNameEntry.Text,
			f.lastNameEntry.Text,
			f.emailEntry.Text,
			f.phoneEntry.Text,
		)

		// Validate
		if err := person.Validate(); err != nil {
			dialog.ShowError(err, f.app.mainWindow)
			return
		}

		// Create in database
		if err := f.app.personCtrl.CreatePerson(f.app.ctx, person); err != nil {
			log.Error().Err(err).Msg("Failed to create person")
			dialog.ShowError(fmt.Errorf("failed to create person: %w", err), f.app.mainWindow)
			return
		}

		log.Info().Str("name", person.FullName()).Msg("Person created successfully")
		dialog.ShowInformation("Success", "Person created successfully", f.app.mainWindow)
	} else {
		// Update existing person
		f.person.FirstName = f.firstNameEntry.Text
		f.person.LastName = f.lastNameEntry.Text

		// Use the domain methods for email and phone to ensure validation
		if err := f.person.UpdateEmail(f.emailEntry.Text); err != nil {
			dialog.ShowError(err, f.app.mainWindow)
			return
		}

		if err := f.person.UpdatePhone(f.phoneEntry.Text); err != nil {
			dialog.ShowError(err, f.app.mainWindow)
			return
		}

		// Validate
		if err := f.person.Validate(); err != nil {
			dialog.ShowError(err, f.app.mainWindow)
			return
		}

		// Update in database
		if err := f.app.personCtrl.UpdatePerson(f.app.ctx, f.person); err != nil {
			log.Error().Err(err).Msg("Failed to update person")
			dialog.ShowError(fmt.Errorf("failed to update person: %w", err), f.app.mainWindow)
			return
		}

		log.Info().Str("name", f.person.FullName()).Msg("Person updated successfully")
		dialog.ShowInformation("Success", "Person updated successfully", f.app.mainWindow)
	}

	// Call onSave callback to refresh the view
	if f.onSave != nil {
		f.onSave()
	}
}
