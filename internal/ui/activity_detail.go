package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// ActivityDetailView displays details of a single activity
type ActivityDetailView struct {
	app                *App
	activityID         int64
	attendees          []*domain.Attendee
	selectedAttendees  map[int64]bool // attendee ID -> selected
	attendeeCheckboxes map[int64]*widget.Check
}

// NewActivityDetailView creates a new activity detail view
func NewActivityDetailView(app *App, activityID int64) *ActivityDetailView {
	return &ActivityDetailView{
		app:                app,
		activityID:         activityID,
		selectedAttendees:  make(map[int64]bool),
		attendeeCheckboxes: make(map[int64]*widget.Check),
	}
}

// Render renders the activity detail view
func (v *ActivityDetailView) Render() fyne.CanvasObject {
	// Get activity
	activity, err := v.app.activityCtrl.GetActivity(v.app.ctx, v.activityID)
	if err != nil {
		log.Error().Err(err).Int64("id", v.activityID).Msg("Failed to load activity")
		return widget.NewLabel("Error loading activity")
	}

	// Back button
	backBtn := widget.NewButton("← Back to Activities", func() {
		v.app.showActivities()
	})

	// Edit button
	editBtn := widget.NewButton("Edit", func() {
		v.showEditActivityDialog(activity)
	})
	editBtn.Importance = widget.HighImportance

	// Header
	header := widget.NewLabelWithStyle(activity.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Header with buttons
	headerContainer := container.NewBorder(nil, nil, backBtn, editBtn, header)

	// Details
	details := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Date: %s", activity.Date.Format("Monday, January 2, 2006"))),
		widget.NewLabel(fmt.Sprintf("Location: %s", activity.Location)),
		widget.NewLabel(fmt.Sprintf("Type: %s", activity.ActivityType)),
		widget.NewLabel(fmt.Sprintf("Status: %s", activity.Status)),
	)

	if activity.Description != "" {
		details.Add(widget.NewSeparator())
		details.Add(widget.NewLabel("Description:"))
		details.Add(widget.NewLabel(activity.Description))
	}

	// Capacity info
	if activity.MaxCapacity != nil {
		_, capacityInfo, err := v.app.activityCtrl.GetActivityWithCapacity(v.app.ctx, v.activityID)
		if err == nil && capacityInfo != nil {
			details.Add(widget.NewSeparator())
			details.Add(widget.NewLabel(fmt.Sprintf("Capacity: %d / %d", capacityInfo.CurrentAttendees, *capacityInfo.MaxCapacity)))
		}
	}

	// Fee info
	if !activity.IsFree {
		feeFloat, _ := activity.Fee.Float64()
		details.Add(widget.NewLabel(fmt.Sprintf("Fee: $%.2f", feeFloat)))
	} else {
		details.Add(widget.NewLabel("Free event"))
	}

	// Payment summary section
	paymentSummary := v.createPaymentSummary()

	// Attendees section
	attendeesHeader := widget.NewLabelWithStyle("Attendees", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Register attendee button
	registerBtn := widget.NewButton("+ Register Attendee", func() {
		v.showRegisterAttendeeDialog()
	})
	registerBtn.Importance = widget.HighImportance

	attendeesHeaderContainer := container.NewBorder(nil, nil, attendeesHeader, registerBtn)
	attendeesList := v.createAttendeesList()
	bulkActions := v.createBulkActions()

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
				paymentSummary,
				widget.NewSeparator(),
				attendeesHeaderContainer,
				bulkActions,
				attendeesList,
			),
		),
	)

	return content
}

// showEditActivityDialog shows dialog to edit activity
func (v *ActivityDetailView) showEditActivityDialog(activity *domain.Activity) {
	form := NewActivityForm(v.app, activity, func() {
		// Refresh the activity detail view after saving
		v.app.showActivityDetail(activity.ID)
	})
	form.Show(v.app.mainWindow)
}

// showRegisterAttendeeDialog shows dialog to register an attendee
func (v *ActivityDetailView) showRegisterAttendeeDialog() {
	form := NewAttendeeForm(v.app, v.activityID, func() {
		// Refresh the activity detail view after registering
		v.app.showActivityDetail(v.activityID)
	})
	form.Show(v.app.mainWindow)
}

// createPaymentSummary creates the payment summary section
func (v *ActivityDetailView) createPaymentSummary() fyne.CanvasObject {
	summary, err := v.app.activityCtrl.GetPaymentSummary(v.app.ctx, v.activityID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load payment summary")
		return widget.NewLabel("Error loading payment summary")
	}

	paidAmount, _ := summary.PaidAmount.Float64()
	unpaidAmount, _ := summary.UnpaidAmount.Float64()

	summaryText := fmt.Sprintf("Payment Summary: %d Paid ($%.2f) | %d Unpaid ($%.2f) | %d Waived",
		summary.PaidCount,
		paidAmount,
		summary.UnpaidCount,
		unpaidAmount,
		summary.WaivedCount)

	return widget.NewLabel(summaryText)
}

// createBulkActions creates the bulk action buttons
func (v *ActivityDetailView) createBulkActions() fyne.CanvasObject {
	markPaidBtn := widget.NewButton("Mark Selected as Paid", func() {
		v.bulkMarkPaid()
	})

	waiveBtn := widget.NewButton("Waive Selected", func() {
		v.bulkWaivePayment()
	})

	selectAllBtn := widget.NewButton("Select All", func() {
		v.selectAll()
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		v.deselectAll()
	})

	return container.NewHBox(
		selectAllBtn,
		deselectAllBtn,
		markPaidBtn,
		waiveBtn,
	)
}

// createAttendeesList creates a list of attendees for this activity with checkboxes
func (v *ActivityDetailView) createAttendeesList() fyne.CanvasObject {
	attendees, err := v.app.activityCtrl.GetAttendees(v.app.ctx, v.activityID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load attendees")
		return widget.NewLabel("Error loading attendees")
	}

	v.attendees = attendees

	if len(attendees) == 0 {
		return widget.NewLabel("No attendees registered yet")
	}

	items := make([]fyne.CanvasObject, 0, len(attendees))
	for _, attendee := range attendees {
		// Get person details to show person name
		person, err := v.app.personCtrl.GetPerson(v.app.ctx, attendee.PersonID)
		if err != nil {
			log.Error().Err(err).Int64("person_id", attendee.PersonID).Msg("Failed to load person")
			continue
		}

		roleText := attendee.GetRoleDisplay()
		statusText := string(attendee.PaymentStatus)

		// Create checkbox for selection
		checkbox := widget.NewCheck("", func(checked bool) {
			v.selectedAttendees[attendee.ID] = checked
		})
		v.attendeeCheckboxes[attendee.ID] = checkbox

		// Create label with attendee info
		label := widget.NewLabel(fmt.Sprintf("%s - %s - %s (Registered: %s)",
			person.FullName(),
			roleText,
			statusText,
			attendee.RegistrationDate.Format("Jan 2")))

		// Combine checkbox and label
		item := container.NewHBox(checkbox, label)
		items = append(items, item)
	}

	return container.NewVBox(items...)
}

// selectAll selects all attendees
func (v *ActivityDetailView) selectAll() {
	for id, checkbox := range v.attendeeCheckboxes {
		checkbox.SetChecked(true)
		v.selectedAttendees[id] = true
	}
}

// deselectAll deselects all attendees
func (v *ActivityDetailView) deselectAll() {
	for id, checkbox := range v.attendeeCheckboxes {
		checkbox.SetChecked(false)
		v.selectedAttendees[id] = false
	}
}

// bulkMarkPaid marks selected attendees as paid
func (v *ActivityDetailView) bulkMarkPaid() {
	// Get selected attendee IDs
	selectedIDs := v.getSelectedAttendeeIDs()
	if len(selectedIDs) == 0 {
		dialog.ShowInformation("No Selection", "Please select at least one attendee", v.app.mainWindow)
		return
	}

	// Get activity to determine fee amount
	activity, err := v.app.activityCtrl.GetActivity(v.app.ctx, v.activityID)
	if err != nil {
		dialog.ShowError(err, v.app.mainWindow)
		return
	}

	// Show confirmation dialog with amount entry
	amountEntry := widget.NewEntry()
	if !activity.IsFree {
		amountEntry.SetText(activity.Fee.String())
	} else {
		amountEntry.SetText("0.00")
	}

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Mark %d attendees as paid?", len(selectedIDs))),
		widget.NewLabel("Payment amount:"),
		amountEntry,
	)

	dialog.ShowCustomConfirm("Confirm Bulk Payment", "Mark Paid", "Cancel", content,
		func(confirm bool) {
			if !confirm {
				return
			}

			// Parse amount
			amount, err := decimal.NewFromString(amountEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid amount: %w", err), v.app.mainWindow)
				return
			}

			// Execute bulk mark paid
			err = v.app.activityCtrl.BulkMarkAttendeesPaid(v.app.ctx, selectedIDs, amount)
			if err != nil {
				dialog.ShowError(err, v.app.mainWindow)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Marked %d attendees as paid", len(selectedIDs)), v.app.mainWindow)
			v.app.showActivityDetail(v.activityID) // Refresh view
		}, v.app.mainWindow)
}

// bulkWaivePayment waives payment for selected attendees
func (v *ActivityDetailView) bulkWaivePayment() {
	// Get selected attendee IDs
	selectedIDs := v.getSelectedAttendeeIDs()
	if len(selectedIDs) == 0 {
		dialog.ShowInformation("No Selection", "Please select at least one attendee", v.app.mainWindow)
		return
	}

	// Show confirmation dialog
	dialog.ShowConfirm("Confirm Waive Payment",
		fmt.Sprintf("Waive payment for %d attendees?", len(selectedIDs)),
		func(confirm bool) {
			if !confirm {
				return
			}

			// Execute bulk waive payment
			err := v.app.activityCtrl.BulkWaiveAttendeesPayment(v.app.ctx, selectedIDs)
			if err != nil {
				dialog.ShowError(err, v.app.mainWindow)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Waived payment for %d attendees", len(selectedIDs)), v.app.mainWindow)
			v.app.showActivityDetail(v.activityID) // Refresh view
		}, v.app.mainWindow)
}

// getSelectedAttendeeIDs returns the IDs of selected attendees
func (v *ActivityDetailView) getSelectedAttendeeIDs() []int64 {
	var ids []int64
	for id, selected := range v.selectedAttendees {
		if selected {
			ids = append(ids, id)
		}
	}
	return ids
}
