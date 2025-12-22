package excel

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func createTestExcelFile(t *testing.T, headers []string, rows [][]interface{}) []byte {
	t.Helper()

	f := excelize.NewFile()
	sheetName := "Sheet1"

	// Write headers
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		require.NoError(t, err)
		require.NoError(t, f.SetCellValue(sheetName, cell, header))
	}

	// Write rows
	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			require.NoError(t, err)
			require.NoError(t, f.SetCellValue(sheetName, cell, value))
		}
	}

	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	require.NoError(t, f.Close())

	return buf.Bytes()
}

func TestParseImportFile_Success(t *testing.T) {
	headers := []string{"Name", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"}
	rows := [][]interface{}{
		{"John Doe", "Governor", "state", "California", "Democratic", "2020-01-01"},
		{"Jane Smith", "Senator", "federal", "United States", "Republican", "2019-01-01"},
	}

	fileData := createTestExcelFile(t, headers, rows)
	result, err := ParseImportFile(fileData)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify first row
	assert.Equal(t, "John Doe", result[0].Name)
	assert.Equal(t, "Governor", result[0].Position)
	assert.Equal(t, "state", result[0].JurisdictionType)
	assert.Equal(t, "California", result[0].JurisdictionName)
	assert.Equal(t, "Democratic", result[0].Party)
	assert.Equal(t, "2020-01-01", result[0].TermStart)
	assert.Equal(t, 2, result[0].RowNumber)

	// Verify second row
	assert.Equal(t, "Jane Smith", result[1].Name)
	assert.Equal(t, "Senator", result[1].Position)
	assert.Equal(t, "federal", result[1].JurisdictionType)
	assert.Equal(t, "United States", result[1].JurisdictionName)
	assert.Equal(t, "Republican", result[1].Party)
	assert.Equal(t, "2019-01-01", result[1].TermStart)
	assert.Equal(t, 3, result[1].RowNumber)
}

func TestParseImportFile_WithOptionalFields(t *testing.T) {
	headers := []string{
		"Name", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start",
		"Term End", "Photo URL", "Short Bio", "Birth Date",
	}
	rows := [][]interface{}{
		{"John Doe", "Governor", "state", "California", "Democratic", "2020-01-01",
			"2024-01-01", "https://example.com/photo.jpg", "Former mayor", "1970-05-15"},
	}

	fileData := createTestExcelFile(t, headers, rows)
	result, err := ParseImportFile(fileData)

	require.NoError(t, err)
	require.Len(t, result, 1)

	assert.Equal(t, "John Doe", result[0].Name)
	require.NotNil(t, result[0].TermEnd)
	assert.Equal(t, "2024-01-01", *result[0].TermEnd)
	require.NotNil(t, result[0].PhotoURL)
	assert.Equal(t, "https://example.com/photo.jpg", *result[0].PhotoURL)
	require.NotNil(t, result[0].ShortBio)
	assert.Equal(t, "Former mayor", *result[0].ShortBio)
	require.NotNil(t, result[0].BirthDate)
	assert.Equal(t, "1970-05-15", *result[0].BirthDate)
}

func TestParseImportFile_EmptySheet(t *testing.T) {
	f := excelize.NewFile()
	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	require.NoError(t, f.Close())

	// Remove default sheet to create empty file
	f2 := excelize.NewFile()
	f2.DeleteSheet("Sheet1")
	var buf2 bytes.Buffer
	require.NoError(t, f2.Write(&buf2))
	require.NoError(t, f2.Close())

	_, err := ParseImportFile(buf2.Bytes())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "excel file has no sheets")
}

func TestParseImportFile_MissingHeaderRow(t *testing.T) {
	headers := []string{"Name", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"}
	rows := [][]interface{}{} // No data rows

	fileData := createTestExcelFile(t, headers, rows)
	_, err := ParseImportFile(fileData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "excel file must have at least a header row and one data row")
}

func TestParseImportFile_MissingRequiredColumn(t *testing.T) {
	testCases := []struct {
		name          string
		headers       []string
		expectedError string
	}{
		{
			name:          "missing Name",
			headers:       []string{"Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"},
			expectedError: "missing required column: name",
		},
		{
			name:          "missing Position",
			headers:       []string{"Name", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"},
			expectedError: "missing required column: position",
		},
		{
			name:          "missing Jurisdiction Type",
			headers:       []string{"Name", "Position", "Jurisdiction Name", "Party", "Term Start"},
			expectedError: "missing required column: jurisdiction type",
		},
		{
			name:          "missing Term Start",
			headers:       []string{"Name", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party"},
			expectedError: "missing required column: term start",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := [][]interface{}{
				{"John Doe", "Governor", "state", "California", "Democratic"},
			}

			fileData := createTestExcelFile(t, tc.headers, rows)
			_, err := ParseImportFile(fileData)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestParseImportFile_SkipEmptyRows(t *testing.T) {
	headers := []string{"Name", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"}
	rows := [][]interface{}{
		{"John Doe", "Governor", "state", "California", "Democratic", "2020-01-01"},
		{"", "", "", "", "", ""}, // Empty row
		{"Jane Smith", "Senator", "federal", "United States", "Republican", "2019-01-01"},
	}

	fileData := createTestExcelFile(t, headers, rows)
	result, err := ParseImportFile(fileData)

	require.NoError(t, err)
	assert.Len(t, result, 2) // Empty row should be skipped
	assert.Equal(t, "John Doe", result[0].Name)
	assert.Equal(t, "Jane Smith", result[1].Name)
}

func TestParseImportFile_NoValidDataRows(t *testing.T) {
	headers := []string{"Name", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"}
	rows := [][]interface{}{
		{"", "", "", "", "", ""}, // Empty row
	}

	fileData := createTestExcelFile(t, headers, rows)
	_, err := ParseImportFile(fileData)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid data rows found in Excel file")
}

func TestParseImportFile_CaseInsensitiveHeaders(t *testing.T) {
	headers := []string{"NAME", "POSITION", "jurisdiction TYPE", "Jurisdiction Name", "party", "Term Start"}
	rows := [][]interface{}{
		{"John Doe", "Governor", "state", "California", "Democratic", "2020-01-01"},
	}

	fileData := createTestExcelFile(t, headers, rows)
	result, err := ParseImportFile(fileData)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "John Doe", result[0].Name)
}

func TestParseImportFile_TrimWhitespace(t *testing.T) {
	headers := []string{"  Name  ", "Position", "Jurisdiction Type", "Jurisdiction Name", "Party", "Term Start"}
	rows := [][]interface{}{
		{"  John Doe  ", "  Governor  ", "  state  ", "  California  ", "  Democratic  ", "  2020-01-01  "},
	}

	fileData := createTestExcelFile(t, headers, rows)
	result, err := ParseImportFile(fileData)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "John Doe", result[0].Name)
	assert.Equal(t, "Governor", result[0].Position)
	assert.Equal(t, "state", result[0].JurisdictionType)
	assert.Equal(t, "California", result[0].JurisdictionName)
}

func TestGetColumnValue(t *testing.T) {
	row := []string{"value1", "value2", "value3"}
	colMap := map[string]int{
		"col1": 0,
		"col2": 1,
		"col3": 2,
	}

	assert.Equal(t, "value1", GetColumnValue(row, colMap, "col1"))
	assert.Equal(t, "value2", GetColumnValue(row, colMap, "col2"))
	assert.Equal(t, "value3", GetColumnValue(row, colMap, "col3"))
	assert.Equal(t, "", GetColumnValue(row, colMap, "nonexistent"))
}

func TestGetColumnValue_OutOfBounds(t *testing.T) {
	row := []string{"value1"}
	colMap := map[string]int{
		"col1": 0,
		"col2": 5, // Out of bounds
	}

	assert.Equal(t, "value1", GetColumnValue(row, colMap, "col1"))
	assert.Equal(t, "", GetColumnValue(row, colMap, "col2"))
}

func TestGetColumnValuePtr(t *testing.T) {
	row := []string{"value1", "value2", ""}
	colMap := map[string]int{
		"col1": 0,
		"col2": 1,
		"col3": 2,
	}

	val1 := GetColumnValuePtr(row, colMap, "col1")
	require.NotNil(t, val1)
	assert.Equal(t, "value1", *val1)

	val2 := GetColumnValuePtr(row, colMap, "col2")
	require.NotNil(t, val2)
	assert.Equal(t, "value2", *val2)

	val3 := GetColumnValuePtr(row, colMap, "col3")
	assert.Nil(t, val3) // Empty string should return nil

	val4 := GetColumnValuePtr(row, colMap, "nonexistent")
	assert.Nil(t, val4)
}
