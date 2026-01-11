package excel

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/xuri/excelize/v2"
)

// ParseImportFile parses an Excel file and returns import rows
func ParseImportFile(fileData []byte) ([]models.ImportRow, error) {
	// Open Excel file from bytes
	f, err := excelize.OpenReader(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet (Politicians sheet)
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("excel file has no sheets")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("excel file must have at least a header row and one data row")
	}

	// Parse header row to find column indices
	headers := rows[0]
	colMap := make(map[string]int)
	for i, header := range headers {
		colMap[strings.TrimSpace(strings.ToLower(header))] = i
	}

	// Required columns
	requiredCols := []string{"name", "position", "jurisdiction type", "jurisdiction name", "party", "term start"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	// Parse data rows
	var importRows []models.ImportRow
	for i, row := range rows[1:] { // Skip header
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue // Skip empty rows
		}

		importRow := models.ImportRow{
			RowNumber: i + 2, // Excel row number (1-indexed, +1 for header)
		}

		// Required fields
		if idx, ok := colMap["name"]; ok && idx < len(row) {
			importRow.Name = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["position"]; ok && idx < len(row) {
			importRow.Position = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["jurisdiction type"]; ok && idx < len(row) {
			importRow.JurisdictionType = strings.TrimSpace(strings.ToLower(row[idx]))
		}
		if idx, ok := colMap["jurisdiction name"]; ok && idx < len(row) {
			importRow.JurisdictionName = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["party"]; ok && idx < len(row) {
			importRow.Party = strings.TrimSpace(row[idx])
		}
		if idx, ok := colMap["term start"]; ok && idx < len(row) {
			importRow.TermStart = strings.TrimSpace(row[idx])
		}

		// Optional fields
		if idx, ok := colMap["term end"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				importRow.TermEnd = &val
			}
		}
		if idx, ok := colMap["photo url"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				importRow.PhotoURL = &val
			}
		}
		if idx, ok := colMap["short bio"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				importRow.ShortBio = &val
			}
		}
		if idx, ok := colMap["birth date"]; ok && idx < len(row) {
			val := strings.TrimSpace(row[idx])
			if val != "" {
				importRow.BirthDate = &val
			}
		}

		importRows = append(importRows, importRow)
	}

	if len(importRows) == 0 {
		return nil, fmt.Errorf("no valid data rows found in Excel file")
	}

	return importRows, nil
}

// GetColumnIndex safely gets column value by name
func GetColumnValue(row []string, colMap map[string]int, colName string) string {
	if idx, ok := colMap[colName]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

// GetColumnValuePtr gets column value as pointer (for optional fields)
func GetColumnValuePtr(row []string, colMap map[string]int, colName string) *string {
	val := GetColumnValue(row, colMap, colName)
	if val == "" {
		return nil
	}
	return &val
}
