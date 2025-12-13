package ui

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// ActivityForm represents a form for creating or editing activities
type ActivityForm struct {
	app      *App
	activity *domain.Activity // nil for create, populated for edit
	onSave   func()

	// Form fields
	nameEntry        *widget.Entry
	descriptionEntry *widget.Entry
	dateEntry        *widget.Entry
	locationEntry    *widget.Entry
	typeSelect       *widget.Select
	statusSelect     *widget.Select
	requiresRegCheck *widget.Check
	isFreeCheck      *widget.Check
	feeEntry         *widget.Entry
	capacityEntry    *widget.Entry
	unlimitedCheck   *widget.Check

	// Form container
	formContainer *fyne.Container
}

// NewActivityForm creates a new activity form
func NewActivityForm(app *App, activity *domain.Activity, onSave func()) *ActivityForm {
	form := &ActivityForm{
		app:      app,
		activity: activity,
		onSave:   onSave,
	}

	form.initFields()
	form.buildForm()

	return form
}

// initFields initializes all form fields
func (f *ActivityForm) initFields() {
	// Name
	f.nameEntry = widget.NewEntry()
	f.nameEntry.SetPlaceHolder("Enter activity name")

	// Description
	f.descriptionEntry = widget.NewMultiLineEntry()
	f.descriptionEntry.SetPlaceHolder("Enter description (optional)")

	// Date
	f.dateEntry = widget.NewEntry()
	f.dateEntry.SetPlaceHolder("YYYY-MM-DD")

	// Location
	f.locationEntry = widget.NewEntry()
	f.locationEntry.SetPlaceHolder("Enter location")

	// Activity Type
	f.typeSelect = widget.NewSelect(
		[]string{"event", "gathering"},
		func(value string) {
			// When type changes, update form visibility
			f.updateFormVisibility()
		},
	)

	// Status
	f.statusSelect = widget.NewSelect(
		[]string{"active", "cancelled", "completed"},
		nil,
	)

	// Requires Registration
	f.requiresRegCheck = widget.NewCheck("Requires Registration", func(checked bool) {
		f.updateFormVisibility()
	})

	// Is Free
	f.isFreeCheck = widget.NewCheck("Free Event", func(checked bool) {
		f.feeEntry.Disable()
		if !checked {
			f.feeEntry.Enable()
		}
	})

	// Fee
	f.feeEntry = widget.NewEntry()
	f.feeEntry.SetPlaceHolder("0.00")

	// Capacity
	f.capacityEntry = widget.NewEntry()
	f.capacityEntry.SetPlaceHolder("Enter max capacity")

	// Unlimited capacity
	f.unlimitedCheck = widget.NewCheck("Unlimited Capacity", func(checked bool) {
		f.capacityEntry.Disable()
		if !checked {
			f.capacityEntry.Enable()
		}
	})

	// Populate fields if editing
	if f.activity != nil {
		f.populateFields()
	} else {
		// Set defaults for new activity
		f.typeSelect.SetSelected("event")
		f.statusSelect.SetSelected("active")
		f.requiresRegCheck.SetChecked(true)
		f.isFreeCheck.SetChecked(false)
		f.dateEntry.SetText(time.Now().Format("2006-01-02"))
	}
}

// populateFields populates form fields from the activity
func (f *ActivityForm) populateFields() {
	f.nameEntry.SetText(f.activity.Name)
	f.descriptionEntry.SetText(f.activity.Description)
	f.dateEntry.SetText(f.activity.Date.Format("2006-01-02"))
	f.locationEntry.SetText(f.activity.Location)
	f.typeSelect.SetSelected(string(f.activity.ActivityType))
	f.statusSelect.SetSelected(string(f.activity.Status))
	f.requiresRegCheck.SetChecked(f.activity.RequiresRegistration)
	f.isFreeCheck.SetChecked(f.activity.IsFree)

	if !f.activity.Fee.IsZero() {
		f.feeEntry.SetText(f.activity.Fee.String())
	}

	if f.activity.MaxCapacity != nil {
		f.capacityEntry.SetText(fmt.Sprintf("%d", *f.activity.MaxCapacity))
		f.unlimitedCheck.SetChecked(false)
	} else {
		f.unlimitedCheck.SetChecked(true)
	}
}

// buildForm builds the form UI
func (f *ActivityForm) buildForm() {
	f.formContainer = container.NewVBox(
		// Name
		widget.NewLabel("Name *"),
		f.nameEntry,

		// Description
		widget.NewLabel("Description"),
		f.descriptionEntry,

		// Date
		widget.NewLabel("Date *"),
		f.dateEntry,

		// Location
		widget.NewLabel("Location *"),
		f.locationEntry,

		// Activity Type
		widget.NewLabel("Activity Type *"),
		f.typeSelect,

		// Status
		widget.NewLabel("Status *"),
		f.statusSelect,

		// Registration settings (only for events)
		f.requiresRegCheck,

		// Fee settings
		f.isFreeCheck,
		widget.NewLabel("Fee"),
		f.feeEntry,

		// Capacity
		f.unlimitedCheck,
		widget.NewLabel("Max Capacity"),
		f.capacityEntry,
	)

	f.updateFormVisibility()
}

// updateFormVisibility updates form field visibility based on selections
func (f *ActivityForm) updateFormVisibility() {
	isEvent := f.typeSelect.Selected == "event"

	// Registration is only for events
	if !isEvent {
		f.requiresRegCheck.SetChecked(false)
	}

	// Fee is only relevant if requires registration and not free
	if f.isFreeCheck.Checked {
		f.feeEntry.Disable()
	} else {
		f.feeEntry.Enable()
	}

	// Capacity is only relevant if not unlimited
	if f.unlimitedCheck.Checked {
		f.capacityEntry.Disable()
	} else {
		f.capacityEntry.Enable()
	}

	f.formContainer.Refresh()
}

// Render renders the form in a dialog
func (f *ActivityForm) Show(parentWindow fyne.Window) {
	title := "New Activity"
	if f.activity != nil {
		title = "Edit Activity"
	}

	// Create scrollable form
	scrollForm := container.NewVScroll(f.formContainer)
	scrollForm.SetMinSize(fyne.NewSize(500, 600))

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

	formDialog.Resize(fyne.NewSize(550, 700))
	formDialog.Show()
}

// handleSave validates and saves the activity
func (f *ActivityForm) handleSave() {
	// Validate required fields
	if f.nameEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("name is required"), f.app.mainWindow)
		return
	}

	if f.locationEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("location is required"), f.app.mainWindow)
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", f.dateEntry.Text)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid date format. Use YYYY-MM-DD"), f.app.mainWindow)
		return
	}

	// Parse fee
	var fee decimal.Decimal
	if !f.isFreeCheck.Checked && f.feeEntry.Text != "" {
		fee, err = decimal.NewFromString(f.feeEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid fee amount"), f.app.mainWindow)
			return
		}
		if fee.IsNegative() {
			dialog.ShowError(fmt.Errorf("fee cannot be negative"), f.app.mainWindow)
			return
		}
	}

	// Parse capacity
	var maxCapacity *int
	if !f.unlimitedCheck.Checked && f.capacityEntry.Text != "" {
		capacity, err := strconv.Atoi(f.capacityEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid capacity. Must be a number"), f.app.mainWindow)
			return
		}
		if capacity < 0 {
			dialog.ShowError(fmt.Errorf("capacity cannot be negative"), f.app.mainWindow)
			return
		}
		maxCapacity = &capacity
	}

	// Create or update activity
	if f.activity == nil {
		// Create new activity
		activity := domain.NewActivity(f.nameEntry.Text, f.locationEntry.Text, date)
		activity.Description = f.descriptionEntry.Text
		activity.ActivityType = domain.ActivityType(f.typeSelect.Selected)
		activity.Status = domain.ActivityStatus(f.statusSelect.Selected)
		activity.RequiresRegistration = f.requiresRegCheck.Checked
		activity.IsFree = f.isFreeCheck.Checked
		activity.Fee = fee
		activity.MaxCapacity = maxCapacity

		// Validate
		if err := activity.Validate(); err != nil {
			dialog.ShowError(err, f.app.mainWindow)
			return
		}

		// Create in database
		if err := f.app.activityCtrl.CreateActivity(f.app.ctx, activity); err != nil {
			log.Error().Err(err).Msg("Failed to create activity")
			dialog.ShowError(fmt.Errorf("failed to create activity: %w", err), f.app.mainWindow)
			return
		}

		log.Info().Str("name", activity.Name).Msg("Activity created successfully")
		dialog.ShowInformation("Success", "Activity created successfully", f.app.mainWindow)
	} else {
		// Update existing activity
		f.activity.Name = f.nameEntry.Text
		f.activity.Description = f.descriptionEntry.Text
		f.activity.Date = date
		f.activity.Location = f.locationEntry.Text
		f.activity.ActivityType = domain.ActivityType(f.typeSelect.Selected)
		f.activity.Status = domain.ActivityStatus(f.statusSelect.Selected)
		f.activity.RequiresRegistration = f.requiresRegCheck.Checked
		f.activity.IsFree = f.isFreeCheck.Checked
		f.activity.Fee = fee
		f.activity.MaxCapacity = maxCapacity

		// Validate
		if err := f.activity.Validate(); err != nil {
			dialog.ShowError(err, f.app.mainWindow)
			return
		}

		// Update in database
		if err := f.app.activityCtrl.UpdateActivity(f.app.ctx, f.activity); err != nil {
			log.Error().Err(err).Msg("Failed to update activity")
			dialog.ShowError(fmt.Errorf("failed to update activity: %w", err), f.app.mainWindow)
			return
		}

		log.Info().Str("name", f.activity.Name).Msg("Activity updated successfully")
		dialog.ShowInformation("Success", "Activity updated successfully", f.app.mainWindow)
	}

	// Call onSave callback to refresh the view
	if f.onSave != nil {
		f.onSave()
	}
}
