package ui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/PodioSpaz/event-tracker-go/internal/domain"
	"github.com/PodioSpaz/event-tracker-go/internal/util"
	"github.com/rs/zerolog/log"
)

// CSVImportDialog handles CSV file import for people
type CSVImportDialog struct {
	app          *App
	window       fyne.Window
	filePath     string
	headers      []string
	mapping      *util.CSVColumnMapping
	parseResult  *util.CSVParseResult
	onComplete   func()

	// UI widgets
	fileLabel      *widget.Label
	mappingForm    *widget.Form
	previewTable   *widget.Table
	statusLabel    *widget.Label
	importBtn      *widget.Button
	cancelBtn      *widget.Button
}

// NewCSVImportDialog creates a new CSV import dialog
func NewCSVImportDialog(app *App, window fyne.Window, onComplete func()) *CSVImportDialog {
	return &CSVImportDialog{
		app:        app,
		window:     window,
		onComplete: onComplete,
		mapping:    util.DefaultCSVColumnMapping(),
	}
}

// Show displays the CSV import dialog
func (d *CSVImportDialog) Show() {
	// Step 1: File selection
	d.showFileSelectionDialog()
}

// showFileSelectionDialog shows a file picker to select CSV file
func (d *CSVImportDialog) showFileSelectionDialog() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, d.window)
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()

		// Get file path
		d.filePath = reader.URI().Path()

		// Read headers
		headers, err := util.ReadCSVHeaders(d.filePath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to read CSV file: %w", err), d.window)
			return
		}

		d.headers = headers

		// Try to auto-detect column mapping
		d.mapping = util.DetectColumnMapping(headers)

		// Show column mapping dialog
		d.showColumnMappingDialog()

	}, d.window)

	// Set filter to CSV files
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
	fileDialog.Show()
}

// showColumnMappingDialog shows a dialog to map CSV columns
func (d *CSVImportDialog) showColumnMappingDialog() {
	// Create dropdowns for column mapping
	firstNameSelect := widget.NewSelect(d.headers, func(value string) {
		d.mapping.FirstNameColumn = value
	})
	lastNameSelect := widget.NewSelect(d.headers, func(value string) {
		d.mapping.LastNameColumn = value
	})
	emailSelect := widget.NewSelect(d.headers, func(value string) {
		d.mapping.EmailColumn = value
	})
	phoneSelect := widget.NewSelect(d.headers, func(value string) {
		d.mapping.PhoneColumn = value
	})

	// Set detected values
	firstNameSelect.SetSelected(d.mapping.FirstNameColumn)
	lastNameSelect.SetSelected(d.mapping.LastNameColumn)
	emailSelect.SetSelected(d.mapping.EmailColumn)
	phoneSelect.SetSelected(d.mapping.PhoneColumn)

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "First Name Column", Widget: firstNameSelect},
			{Text: "Last Name Column", Widget: lastNameSelect},
			{Text: "Email Column (optional)", Widget: emailSelect},
			{Text: "Phone Column (optional)", Widget: phoneSelect},
		},
		OnSubmit: func() {
			// Validate mapping
			if d.mapping.FirstNameColumn == "" || d.mapping.LastNameColumn == "" {
				dialog.ShowError(fmt.Errorf("First Name and Last Name columns are required"), d.window)
				return
			}

			// Parse the CSV file with the mapping
			d.parseCSVFile()
		},
		OnCancel: func() {
			// User cancelled, do nothing
		},
	}

	// Info label
	infoLabel := widget.NewLabel(fmt.Sprintf("File: %s\nMap CSV columns to person fields:", filepath.Base(d.filePath)))

	content := container.NewVBox(
		infoLabel,
		widget.NewSeparator(),
		form,
	)

	dialog.ShowCustomConfirm("Map CSV Columns", "Next", "Cancel", content, func(proceed bool) {
		if proceed {
			form.OnSubmit()
		}
	}, d.window)
}

// parseCSVFile parses the CSV file with the current mapping
func (d *CSVImportDialog) parseCSVFile() {
	parser := util.NewCSVParser(d.mapping)
	result, err := parser.ParseFile(d.filePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to parse CSV file: %w", err), d.window)
		return
	}

	d.parseResult = result

	// Show preview dialog
	d.showPreviewDialog()
}

// showPreviewDialog shows a preview of the parsed data
func (d *CSVImportDialog) showPreviewDialog() {
	// Create status label
	statusText := fmt.Sprintf("Total: %d | Valid: %d | Errors: %d",
		d.parseResult.TotalRows,
		d.parseResult.ValidCount,
		d.parseResult.ErrorCount)
	d.statusLabel = widget.NewLabel(statusText)

	// Create preview table (show first 10 rows)
	previewRows := d.parseResult.Rows
	if len(previewRows) > 10 {
		previewRows = previewRows[:10]
	}

	d.previewTable = widget.NewTable(
		func() (int, int) {
			return len(previewRows) + 1, 5 // +1 for header, 5 columns: First, Last, Email, Phone, Status
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)

			// Header row
			if id.Row == 0 {
				headers := []string{"First Name", "Last Name", "Email", "Phone", "Status"}
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}

			// Data rows
			rowIdx := id.Row - 1
			if rowIdx >= len(previewRows) {
				label.SetText("")
				return
			}

			row := previewRows[rowIdx]
			switch id.Col {
			case 0:
				label.SetText(row.FirstName)
			case 1:
				label.SetText(row.LastName)
			case 2:
				label.SetText(row.Email)
			case 3:
				label.SetText(row.Phone)
			case 4:
				if row.HasErrors() {
					label.SetText("❌ Error")
				} else {
					label.SetText("✓ Valid")
				}
			}
		},
	)

	// Set column widths
	d.previewTable.SetColumnWidth(0, 120)
	d.previewTable.SetColumnWidth(1, 120)
	d.previewTable.SetColumnWidth(2, 180)
	d.previewTable.SetColumnWidth(3, 120)
	d.previewTable.SetColumnWidth(4, 80)

	// Create error details if there are errors
	var errorDetails *widget.Label
	if d.parseResult.ErrorCount > 0 {
		errorText := "Rows with errors:\n"
		for i, row := range d.parseResult.InvalidRows {
			if i < 5 { // Show first 5 errors
				errorText += fmt.Sprintf("Line %d: %s\n", row.LineNumber, row.ErrorString())
			} else {
				errorText += fmt.Sprintf("... and %d more errors\n", len(d.parseResult.InvalidRows)-5)
				break
			}
		}
		errorDetails = widget.NewLabel(errorText)
	}

	// Import and Cancel buttons
	d.importBtn = widget.NewButton("Import Valid Rows", func() {
		d.importPeople()
	})
	d.importBtn.Importance = widget.HighImportance
	if d.parseResult.ValidCount == 0 {
		d.importBtn.Disable()
	}

	d.cancelBtn = widget.NewButton("Cancel", func() {
		// Close dialog
	})

	buttons := container.NewHBox(d.importBtn, d.cancelBtn)

	// Build content
	var content *fyne.Container
	if errorDetails != nil {
		content = container.NewBorder(
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Preview (showing first 10 of %d rows):", len(d.parseResult.Rows))),
				d.statusLabel,
				widget.NewSeparator(),
			),
			container.NewVBox(
				widget.NewSeparator(),
				errorDetails,
				widget.NewSeparator(),
				buttons,
			),
			nil,
			nil,
			d.previewTable,
		)
	} else {
		content = container.NewBorder(
			container.NewVBox(
				widget.NewLabel(fmt.Sprintf("Preview (showing first 10 of %d rows):", len(d.parseResult.Rows))),
				d.statusLabel,
				widget.NewSeparator(),
			),
			container.NewVBox(
				widget.NewSeparator(),
				buttons,
			),
			nil,
			nil,
			d.previewTable,
		)
	}

	// Create dialog
	customDialog := dialog.NewCustom("Preview CSV Import", "Close", content, d.window)
	customDialog.Resize(fyne.NewSize(700, 500))

	d.cancelBtn.OnTapped = func() {
		customDialog.Hide()
	}

	customDialog.Show()
}

// importPeople imports the valid people into the database
func (d *CSVImportDialog) importPeople() {
	if d.parseResult.ValidCount == 0 {
		dialog.ShowInformation("No Data", "No valid rows to import", d.window)
		return
	}

	// Disable import button during import
	d.importBtn.Disable()
	d.statusLabel.SetText(fmt.Sprintf("Importing %d people...", d.parseResult.ValidCount))

	// Convert valid CSV rows to domain.Person objects
	people := make([]*domain.Person, 0, len(d.parseResult.ValidRows))
	for _, row := range d.parseResult.ValidRows {
		person := domain.NewPerson(row.FirstName, row.LastName, row.Email, row.Phone)
		people = append(people, person)
	}

	// Import using BulkCreatePeople
	err := d.app.personCtrl.BulkCreatePeople(d.app.ctx, people)
	if err != nil {
		log.Error().Err(err).Msg("Failed to import people")
		dialog.ShowError(fmt.Errorf("failed to import people: %w", err), d.window)
		d.importBtn.Enable()
		d.statusLabel.SetText(fmt.Sprintf("Import failed: %v", err))
		return
	}

	// Success
	log.Info().Msgf("Successfully imported %d people", len(people))
	dialog.ShowInformation("Success",
		fmt.Sprintf("Successfully imported %d people", len(people)),
		d.window)

	// Call completion callback
	if d.onComplete != nil {
		d.onComplete()
	}
}
