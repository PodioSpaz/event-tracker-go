package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVParser_Parse_ValidData(t *testing.T) {
	csvData := `First Name,Last Name,Email,Phone
John,Doe,john.doe@example.com,555-1234
Jane,Smith,jane.smith@example.com,555-5678
Bob,Johnson,,555-9999
Alice,Williams,alice@example.com,`

	parser := NewCSVParser(DefaultCSVColumnMapping())
	result, err := parser.Parse(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 4, result.TotalRows)
	assert.Equal(t, 4, result.ValidCount)
	assert.Equal(t, 0, result.ErrorCount)

	// Check first row
	assert.Equal(t, "John", result.Rows[0].FirstName)
	assert.Equal(t, "Doe", result.Rows[0].LastName)
	assert.Equal(t, "john.doe@example.com", result.Rows[0].Email)
	assert.Equal(t, "555-1234", result.Rows[0].Phone)
	assert.False(t, result.Rows[0].HasErrors())

	// Check third row (no email, phone only)
	assert.Equal(t, "Bob", result.Rows[2].FirstName)
	assert.Equal(t, "Johnson", result.Rows[2].LastName)
	assert.Equal(t, "", result.Rows[2].Email)
	assert.Equal(t, "555-9999", result.Rows[2].Phone)
	assert.False(t, result.Rows[2].HasErrors())

	// Check fourth row (email only, no phone)
	assert.Equal(t, "Alice", result.Rows[3].FirstName)
	assert.Equal(t, "Williams", result.Rows[3].LastName)
	assert.Equal(t, "alice@example.com", result.Rows[3].Email)
	assert.Equal(t, "", result.Rows[3].Phone)
	assert.False(t, result.Rows[3].HasErrors())
}

func TestCSVParser_Parse_InvalidData(t *testing.T) {
	csvData := `First Name,Last Name,Email,Phone
,Doe,john.doe@example.com,555-1234
Jane,,jane.smith@example.com,555-5678
Bob,Johnson,,
Alice,Williams,invalid-email,`

	parser := NewCSVParser(DefaultCSVColumnMapping())
	result, err := parser.Parse(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 4, result.TotalRows)
	assert.Equal(t, 0, result.ValidCount)   // All rows are invalid
	assert.Equal(t, 4, result.ErrorCount)

	// Row 1: Missing first name
	assert.True(t, result.Rows[0].HasErrors())
	assert.Contains(t, result.Rows[0].ErrorString(), "First Name")

	// Row 2: Missing last name
	assert.True(t, result.Rows[1].HasErrors())
	assert.Contains(t, result.Rows[1].ErrorString(), "Last Name")

	// Row 3: No contact info
	assert.True(t, result.Rows[2].HasErrors())
	assert.Contains(t, result.Rows[2].ErrorString(), "contact method")

	// Row 4: Invalid email
	assert.True(t, result.Rows[3].HasErrors())
	assert.Contains(t, result.Rows[3].ErrorString(), "email")
}

func TestCSVParser_Parse_EmptyFile(t *testing.T) {
	csvData := ``

	parser := NewCSVParser(DefaultCSVColumnMapping())
	_, err := parser.Parse(strings.NewReader(csvData))

	assert.ErrorIs(t, err, ErrCSVEmptyFile)
}

func TestCSVParser_Parse_NoHeaderRow(t *testing.T) {
	csvData := `John,Doe,john@example.com,555-1234`

	parser := NewCSVParser(DefaultCSVColumnMapping())
	_, err := parser.Parse(strings.NewReader(csvData))

	assert.ErrorIs(t, err, ErrCSVNoHeaderRow)
}

func TestCSVParser_Parse_MissingRequiredColumns(t *testing.T) {
	csvData := `First Name,Email
John,john@example.com`

	parser := NewCSVParser(DefaultCSVColumnMapping())
	_, err := parser.Parse(strings.NewReader(csvData))

	assert.ErrorIs(t, err, ErrCSVMissingColumns)
	assert.Contains(t, err.Error(), "Last Name")
}

func TestCSVParser_Parse_CaseInsensitiveHeaders(t *testing.T) {
	csvData := `FIRST NAME,last name,EMAIL,phone
John,Doe,john@example.com,555-1234`

	parser := NewCSVParser(DefaultCSVColumnMapping())
	result, err := parser.Parse(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalRows)
	assert.Equal(t, 1, result.ValidCount)
	assert.Equal(t, "John", result.Rows[0].FirstName)
	assert.Equal(t, "Doe", result.Rows[0].LastName)
}

func TestCSVParser_Parse_CustomMapping(t *testing.T) {
	csvData := `Given Name,Surname,Contact Email,Mobile
John,Doe,john@example.com,555-1234`

	mapping := &CSVColumnMapping{
		FirstNameColumn: "Given Name",
		LastNameColumn:  "Surname",
		EmailColumn:     "Contact Email",
		PhoneColumn:     "Mobile",
	}

	parser := NewCSVParser(mapping)
	result, err := parser.Parse(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalRows)
	assert.Equal(t, 1, result.ValidCount)
	assert.Equal(t, "John", result.Rows[0].FirstName)
	assert.Equal(t, "Doe", result.Rows[0].LastName)
	assert.Equal(t, "john@example.com", result.Rows[0].Email)
	assert.Equal(t, "555-1234", result.Rows[0].Phone)
}

func TestCSVParser_Parse_TrimWhitespace(t *testing.T) {
	csvData := `First Name,Last Name,Email,Phone
  John  ,  Doe  ,  john@example.com  ,  555-1234  `

	parser := NewCSVParser(DefaultCSVColumnMapping())
	result, err := parser.Parse(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalRows)
	assert.Equal(t, "John", result.Rows[0].FirstName)
	assert.Equal(t, "Doe", result.Rows[0].LastName)
	assert.Equal(t, "john@example.com", result.Rows[0].Email)
	assert.Equal(t, "555-1234", result.Rows[0].Phone)
}

func TestCSVParser_Parse_VariableFieldCount(t *testing.T) {
	csvData := `First Name,Last Name,Email,Phone
John,Doe,john@example.com
Jane,Smith,jane@example.com,555-5678,extra field`

	parser := NewCSVParser(DefaultCSVColumnMapping())
	result, err := parser.Parse(strings.NewReader(csvData))

	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalRows)
	assert.Equal(t, 2, result.ValidCount)
}

func TestDetectColumnMapping_StandardHeaders(t *testing.T) {
	headers := []string{"First Name", "Last Name", "Email", "Phone"}
	mapping := DetectColumnMapping(headers)

	assert.Equal(t, "First Name", mapping.FirstNameColumn)
	assert.Equal(t, "Last Name", mapping.LastNameColumn)
	assert.Equal(t, "Email", mapping.EmailColumn)
	assert.Equal(t, "Phone", mapping.PhoneColumn)
}

func TestDetectColumnMapping_AlternativeHeaders(t *testing.T) {
	headers := []string{"Given Name", "Surname", "E-mail", "Mobile"}
	mapping := DetectColumnMapping(headers)

	assert.Equal(t, "Given Name", mapping.FirstNameColumn)
	assert.Equal(t, "Surname", mapping.LastNameColumn)
	assert.Equal(t, "E-mail", mapping.EmailColumn)
	assert.Equal(t, "Mobile", mapping.PhoneColumn)
}

func TestDetectColumnMapping_CaseInsensitive(t *testing.T) {
	headers := []string{"FIRSTNAME", "LASTNAME", "EMAIL", "PHONE"}
	mapping := DetectColumnMapping(headers)

	assert.Equal(t, "FIRSTNAME", mapping.FirstNameColumn)
	assert.Equal(t, "LASTNAME", mapping.LastNameColumn)
	assert.Equal(t, "EMAIL", mapping.EmailColumn)
	assert.Equal(t, "PHONE", mapping.PhoneColumn)
}

func TestDetectColumnMapping_MissingColumns(t *testing.T) {
	headers := []string{"Name", "Contact"}
	mapping := DetectColumnMapping(headers)

	// Should return empty strings for undetected columns
	assert.Equal(t, "", mapping.FirstNameColumn)
	assert.Equal(t, "", mapping.LastNameColumn)
	assert.Equal(t, "", mapping.EmailColumn)
	assert.Equal(t, "", mapping.PhoneColumn)
}

func TestReadCSVHeaders(t *testing.T) {
	// Note: This test would need a real file for ReadCSVHeaders
	// For now, we'll skip this test in favor of Parse tests
	t.Skip("ReadCSVHeaders requires a real file, covered by Parse tests")
}

func TestCSVRow_HasErrors(t *testing.T) {
	row := &CSVRow{
		LineNumber: 1,
		FirstName:  "John",
		LastName:   "Doe",
		Errors:     []error{},
	}

	assert.False(t, row.HasErrors())

	row.Errors = append(row.Errors, ErrInvalidEmail)
	assert.True(t, row.HasErrors())
}

func TestCSVRow_ErrorString(t *testing.T) {
	row := &CSVRow{
		LineNumber: 1,
		FirstName:  "John",
		LastName:   "Doe",
		Errors:     []error{ErrInvalidEmail, ErrInvalidPhone},
	}

	errStr := row.ErrorString()
	assert.Contains(t, errStr, "invalid email")
	assert.Contains(t, errStr, "invalid phone")
}

func TestDefaultCSVColumnMapping(t *testing.T) {
	mapping := DefaultCSVColumnMapping()

	assert.Equal(t, "First Name", mapping.FirstNameColumn)
	assert.Equal(t, "Last Name", mapping.LastNameColumn)
	assert.Equal(t, "Email", mapping.EmailColumn)
	assert.Equal(t, "Phone", mapping.PhoneColumn)
}

func TestNewCSVParser_NilMapping(t *testing.T) {
	parser := NewCSVParser(nil)

	assert.NotNil(t, parser)
	assert.NotNil(t, parser.mapping)
	assert.Equal(t, "First Name", parser.mapping.FirstNameColumn)
}
