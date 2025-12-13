package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/rs/zerolog/log"
)

// PeopleView displays the list of people
type PeopleView struct {
	app          *App
	people       []*domain.Person
	filteredPeople []*domain.Person
	searchEntry  *widget.Entry
	list         *widget.List
}

// NewPeopleView creates a new people view
func NewPeopleView(app *App) *PeopleView {
	return &PeopleView{
		app: app,
		people: make([]*domain.Person, 0),
		filteredPeople: make([]*domain.Person, 0),
	}
}

// Render renders the people list view
func (v *PeopleView) Render() fyne.CanvasObject {
	// Load people
	if err := v.loadPeople(); err != nil {
		log.Error().Err(err).Msg("Failed to load people")
		return widget.NewLabel("Error loading people")
	}

	// Header with add and import buttons
	header := widget.NewLabelWithStyle("People", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	addBtn := widget.NewButton("New Person", func() {
		v.showNewPersonDialog()
	})
	addBtn.Importance = widget.HighImportance

	importBtn := widget.NewButton("Import CSV", func() {
		v.showImportDialog()
	})

	buttonsContainer := container.NewHBox(importBtn, addBtn)
	headerContainer := container.NewBorder(nil, nil, header, buttonsContainer)

	// Search entry
	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("Search by name or email...")
	v.searchEntry.OnChanged = func(query string) {
		v.filterPeople(query)
	}

	// Create list
	v.list = widget.NewList(
		func() int {
			return len(v.filteredPeople)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(v.filteredPeople) {
				person := v.filteredPeople[id]
				contactInfo := person.Email
				if contactInfo == "" {
					contactInfo = person.Phone
				}
				text := fmt.Sprintf("%s (%s)", person.FullName(), contactInfo)
				item.(*fyne.Container).Objects[0].(*widget.Label).SetText(text)
			}
		},
	)

	v.list.OnSelected = func(id widget.ListItemID) {
		if id < len(v.filteredPeople) {
			v.app.showPersonDetail(v.filteredPeople[id].ID)
		}
	}

	// Main content
	content := container.NewBorder(
		container.NewVBox(
			headerContainer,
			widget.NewSeparator(),
			v.searchEntry,
		),
		nil,
		nil,
		nil,
		v.list,
	)

	return content
}

// loadPeople loads all people from the database
func (v *PeopleView) loadPeople() error {
	people, err := v.app.personCtrl.GetAllPeople(v.app.ctx)
	if err != nil {
		return err
	}

	v.people = people
	v.filteredPeople = people
	return nil
}

// filterPeople filters the people list based on the search query
func (v *PeopleView) filterPeople(query string) {
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		// No filter, show all people
		v.filteredPeople = v.people
	} else {
		// Filter people by name or email
		filtered := make([]*domain.Person, 0)
		for _, person := range v.people {
			// Search in first name, last name, full name, email, and phone
			if strings.Contains(strings.ToLower(person.FirstName), query) ||
				strings.Contains(strings.ToLower(person.LastName), query) ||
				strings.Contains(strings.ToLower(person.FullName()), query) ||
				strings.Contains(strings.ToLower(person.Email), query) ||
				strings.Contains(strings.ToLower(person.Phone), query) {
				filtered = append(filtered, person)
			}
		}
		v.filteredPeople = filtered
	}

	// Refresh the list
	if v.list != nil {
		v.list.Refresh()
	}
}

// showNewPersonDialog shows dialog to create new person
func (v *PeopleView) showNewPersonDialog() {
	form := NewPersonForm(v.app, nil, func() {
		// Refresh the people view after saving
		v.app.showPeople()
	})
	form.Show(v.app.mainWindow)
}

// showImportDialog shows CSV import dialog
func (v *PeopleView) showImportDialog() {
	importDialog := NewCSVImportDialog(v.app, v.app.mainWindow, func() {
		// Refresh the people view after import
		v.app.showPeople()
	})
	importDialog.Show()
}
