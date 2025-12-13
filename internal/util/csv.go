package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	// CSV-specific errors
	ErrCSVFileNotFound    = errors.New("CSV file not found")
	ErrCSVParseError      = errors.New("CSV parse error")
	ErrCSVEmptyFile       = errors.New("CSV file is empty")
	ErrCSVNoHeaderRow     = errors.New("CSV file has no header row")
	ErrCSVInvalidHeader   = errors.New("CSV file has invalid headers")
	ErrCSVMissingColumns  = errors.New("CSV file is missing required columns")
	ErrCSVColumnMapping   = errors.New("column mapping error")
)

// CSVColumnMapping defines the mapping between CSV columns and Person fields
type CSVColumnMapping struct {
	FirstNameColumn string // CSV header name for first name
	LastNameColumn  string // CSV header name for last name
	EmailColumn     string // CSV header name for email (optional)
	PhoneColumn     string // CSV header name for phone (optional)
}

// DefaultCSVColumnMapping returns default column mapping
func DefaultCSVColumnMapping() *CSVColumnMapping {
	return &CSVColumnMapping{
		FirstNameColumn: "First Name",
		LastNameColumn:  "Last Name",
		EmailColumn:     "Email",
		PhoneColumn:     "Phone",
	}
}

// CSVRow represents a single row from the CSV file
type CSVRow struct {
	LineNumber int
	FirstName  string
	LastName   string
	Email      string
	Phone      string
	Errors     []error // Validation errors for this row
}

// HasErrors returns true if the row has validation errors
func (r *CSVRow) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorString returns all errors as a single string
func (r *CSVRow) ErrorString() string {
	if !r.HasErrors() {
		return ""
	}

	var errStrs []string
	for _, err := range r.Errors {
		errStrs = append(errStrs, err.Error())
	}
	return strings.Join(errStrs, "; ")
}

// CSVParseResult represents the result of parsing a CSV file
type CSVParseResult struct {
	Rows       []*CSVRow
	ValidRows  []*CSVRow
	InvalidRows []*CSVRow
	TotalRows  int
	ValidCount int
	ErrorCount int
}

// CSVParser handles parsing CSV files for person import
type CSVParser struct {
	mapping *CSVColumnMapping
}

// NewCSVParser creates a new CSV parser with the given column mapping
func NewCSVParser(mapping *CSVColumnMapping) *CSVParser {
	if mapping == nil {
		mapping = DefaultCSVColumnMapping()
	}

	return &CSVParser{
		mapping: mapping,
	}
}

// ParseFile parses a CSV file and returns the result
func (p *CSVParser) ParseFile(filePath string) (*CSVParseResult, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCSVFileNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrCSVParseError, err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses CSV data from an io.Reader
func (p *CSVParser) Parse(reader io.Reader) (*CSVParseResult, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCSVParseError, err)
	}

	if len(records) == 0 {
		return nil, ErrCSVEmptyFile
	}

	if len(records) < 2 {
		return nil, ErrCSVNoHeaderRow
	}

	// Parse header row
	headers := records[0]
	columnIndices, err := p.mapColumns(headers)
	if err != nil {
		return nil, err
	}

	// Parse data rows
	result := &CSVParseResult{
		Rows:       make([]*CSVRow, 0, len(records)-1),
		ValidRows:  make([]*CSVRow, 0),
		InvalidRows: make([]*CSVRow, 0),
	}

	for i := 1; i < len(records); i++ {
		record := records[i]
		lineNumber := i + 1 // Line number in file (1-based, accounting for header)

		row := p.parseRow(record, columnIndices, lineNumber)
		result.Rows = append(result.Rows, row)
		result.TotalRows++

		if row.HasErrors() {
			result.InvalidRows = append(result.InvalidRows, row)
			result.ErrorCount++
		} else {
			result.ValidRows = append(result.ValidRows, row)
			result.ValidCount++
		}
	}

	return result, nil
}

// mapColumns finds the column indices for each required field
func (p *CSVParser) mapColumns(headers []string) (map[string]int, error) {
	indices := make(map[string]int)

	// Normalize headers (trim spaces, case-insensitive)
	normalizedHeaders := make([]string, len(headers))
	for i, h := range headers {
		normalizedHeaders[i] = strings.TrimSpace(h)
	}

	// Find column indices
	firstNameIdx := p.findColumnIndex(normalizedHeaders, p.mapping.FirstNameColumn)
	lastNameIdx := p.findColumnIndex(normalizedHeaders, p.mapping.LastNameColumn)
	emailIdx := p.findColumnIndex(normalizedHeaders, p.mapping.EmailColumn)
	phoneIdx := p.findColumnIndex(normalizedHeaders, p.mapping.PhoneColumn)

	// First name and last name are required
	if firstNameIdx == -1 {
		return nil, fmt.Errorf("%w: missing '%s' column", ErrCSVMissingColumns, p.mapping.FirstNameColumn)
	}
	if lastNameIdx == -1 {
		return nil, fmt.Errorf("%w: missing '%s' column", ErrCSVMissingColumns, p.mapping.LastNameColumn)
	}

	indices["first_name"] = firstNameIdx
	indices["last_name"] = lastNameIdx

	if emailIdx != -1 {
		indices["email"] = emailIdx
	}
	if phoneIdx != -1 {
		indices["phone"] = phoneIdx
	}

	return indices, nil
}

// findColumnIndex finds the index of a column by name (case-insensitive)
func (p *CSVParser) findColumnIndex(headers []string, columnName string) int {
	normalizedName := strings.TrimSpace(columnName)

	for i, h := range headers {
		if strings.EqualFold(h, normalizedName) {
			return i
		}
	}

	return -1
}

// parseRow parses a single CSV row into a CSVRow
func (p *CSVParser) parseRow(record []string, columnIndices map[string]int, lineNumber int) *CSVRow {
	row := &CSVRow{
		LineNumber: lineNumber,
		Errors:     make([]error, 0),
	}

	// Extract values from record
	row.FirstName = p.getFieldValue(record, columnIndices["first_name"])
	row.LastName = p.getFieldValue(record, columnIndices["last_name"])

	if idx, ok := columnIndices["email"]; ok {
		row.Email = p.getFieldValue(record, idx)
	}
	if idx, ok := columnIndices["phone"]; ok {
		row.Phone = p.getFieldValue(record, idx)
	}

	// Validate the row
	p.validateRow(row)

	return row
}

// getFieldValue safely gets a field value from a record
func (p *CSVParser) getFieldValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

// validateRow validates a CSV row
func (p *CSVParser) validateRow(row *CSVRow) {
	// Required fields
	if err := ValidateRequired(row.FirstName, "First Name"); err != nil {
		row.Errors = append(row.Errors, err)
	}
	if err := ValidateRequired(row.LastName, "Last Name"); err != nil {
		row.Errors = append(row.Errors, err)
	}

	// At least one contact method required
	hasEmail := strings.TrimSpace(row.Email) != ""
	hasPhone := strings.TrimSpace(row.Phone) != ""

	if !hasEmail && !hasPhone {
		row.Errors = append(row.Errors, ErrPersonContactRequired)
	}

	// Validate email format if provided
	if hasEmail {
		if err := ValidateEmail(row.Email); err != nil {
			row.Errors = append(row.Errors, fmt.Errorf("email: %w", err))
		}
	}

	// Validate phone format if provided
	if hasPhone {
		if err := ValidatePhone(row.Phone); err != nil {
			row.Errors = append(row.Errors, fmt.Errorf("phone: %w", err))
		}
	}
}

// DetectColumnMapping attempts to auto-detect column mapping from CSV headers
func DetectColumnMapping(headers []string) *CSVColumnMapping {
	mapping := &CSVColumnMapping{}

	// Normalize headers
	normalizedHeaders := make([]string, len(headers))
	for i, h := range headers {
		normalizedHeaders[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Common patterns for first name
	firstNamePatterns := []string{"first name", "firstname", "first", "given name", "givenname"}
	mapping.FirstNameColumn = detectColumn(headers, normalizedHeaders, firstNamePatterns)

	// Common patterns for last name
	lastNamePatterns := []string{"last name", "lastname", "last", "surname", "family name", "familyname"}
	mapping.LastNameColumn = detectColumn(headers, normalizedHeaders, lastNamePatterns)

	// Common patterns for email
	emailPatterns := []string{"email", "email address", "e-mail", "mail"}
	mapping.EmailColumn = detectColumn(headers, normalizedHeaders, emailPatterns)

	// Common patterns for phone
	phonePatterns := []string{"phone", "phone number", "telephone", "tel", "mobile", "cell"}
	mapping.PhoneColumn = detectColumn(headers, normalizedHeaders, phonePatterns)

	return mapping
}

// detectColumn finds the original header that matches one of the patterns
func detectColumn(originalHeaders, normalizedHeaders []string, patterns []string) string {
	for i, normalized := range normalizedHeaders {
		for _, pattern := range patterns {
			if normalized == pattern {
				return originalHeaders[i]
			}
		}
	}
	return ""
}

// ReadCSVHeaders reads just the header row from a CSV file
func ReadCSVHeaders(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCSVFileNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrCSVParseError, err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	headers, err := csvReader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, ErrCSVEmptyFile
		}
		return nil, fmt.Errorf("%w: %v", ErrCSVParseError, err)
	}

	// Trim spaces from headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	return headers, nil
}
