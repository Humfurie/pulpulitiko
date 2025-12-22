package excel

import (
	"fmt"
	"time"

	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/xuri/excelize/v2"
)

// GenerateExportFile generates an Excel file from politician data
func GenerateExportFile(politicians []models.Politician) (*excelize.File, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Politicians"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1") // Remove default sheet

	// Define headers
	headers := []string{
		"Name", "Position", "Jurisdiction Type", "Jurisdiction Name",
		"Party", "Term Start", "Term End", "Photo URL", "Short Bio",
	}

	// Write headers
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Style headers
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	_ = f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", string(rune('A'+len(headers)-1))), headerStyle)

	// Write data
	for i, politician := range politicians {
		row := i + 2 // Start from row 2 (after header)

		// Get position and jurisdiction info (these would come from joined data)
		positionName := ""
		jurisdictionType := ""
		jurisdictionName := ""
		partyName := ""
		termStart := ""
		termEnd := ""

		if politician.PositionInfo != nil {
			positionName = politician.PositionInfo.Name
		}
		if politician.PartyInfo != nil {
			partyName = politician.PartyInfo.Name
		}

		// Format dates
		if politician.TermStart != nil {
			termStart = politician.TermStart.Format("2006-01-02")
		}
		if politician.TermEnd != nil {
			termEnd = politician.TermEnd.Format("2006-01-02")
		}

		// Write row data
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), politician.Name)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), positionName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), jurisdictionType)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), jurisdictionName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), partyName)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), termStart)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), termEnd)
		if politician.Photo != nil {
			_ = f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), *politician.Photo)
		}
		if politician.ShortBio != nil {
			_ = f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), *politician.ShortBio)
		}
	}

	// Auto-fit columns
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		_ = f.SetColWidth(sheetName, col, col, 20)
	}

	return f, nil
}

// GenerateTemplateFile generates an Excel template with data validation
func GenerateTemplateFile(positions []models.GovernmentPosition, parties []models.PoliticalParty, locations map[string][]string) (*excelize.File, error) {
	f := excelize.NewFile()

	// Sheet 1: Politicians (Main template)
	sheetName := "Politicians"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1")

	// Headers with instructions
	headers := []struct {
		Name     string
		Required bool
		Example  string
	}{
		{"Name", true, "Juan Dela Cruz"},
		{"Position", true, "City Mayor"},
		{"Jurisdiction Type", true, "city"},
		{"Jurisdiction Name", true, "Makati City"},
		{"Party", true, "PDP-LABAN"},
		{"Term Start", true, "2022-06-30"},
		{"Term End", false, "2025-06-30"},
		{"Photo URL", false, "https://example.com/photo.jpg"},
		{"Short Bio", false, "Brief biography..."},
	}

	// Write headers
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		headerText := header.Name
		if header.Required {
			headerText += " *"
		}
		_ = f.SetCellValue(sheetName, cell, headerText)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Write example row
	exampleStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E7E6E6"}, Pattern: 1},
		Font: &excelize.Font{Italic: true},
	})

	for i, header := range headers {
		cell := fmt.Sprintf("%s2", string(rune('A'+i)))
		_ = f.SetCellValue(sheetName, cell, header.Example)
		_ = f.SetCellStyle(sheetName, cell, cell, exampleStyle)
	}

	// Add data validation dropdowns
	addPositionValidation(f, sheetName, positions)
	addPartyValidation(f, sheetName, parties)
	addJurisdictionTypeValidation(f, sheetName)

	// Set column widths
	widths := []float64{25, 25, 20, 25, 20, 15, 15, 40, 50}
	for i, width := range widths {
		col := string(rune('A' + i))
		_ = f.SetColWidth(sheetName, col, col, width)
	}

	// Sheet 2: Valid Positions (Reference)
	createReferenceSheet(f, "Valid Positions", positions, func(p models.GovernmentPosition) []string {
		return []string{p.Name, p.Level, p.Branch}
	}, []string{"Position Name", "Level", "Branch"})

	// Sheet 3: Valid Parties (Reference)
	createReferenceSheet(f, "Valid Parties", parties, func(p models.PoliticalParty) []string {
		abbr := ""
		if p.Abbreviation != nil {
			abbr = *p.Abbreviation
		}
		return []string{p.Name, abbr}
	}, []string{"Party Name", "Abbreviation"})

	// Sheet 4: Instructions
	createInstructionsSheet(f)

	return f, nil
}

// Helper functions

func addPositionValidation(f *excelize.File, sheetName string, positions []models.GovernmentPosition) {
	// Create list of position names for validation
	positionNames := make([]string, len(positions))
	for i, p := range positions {
		positionNames[i] = p.Name
	}

	// Add dropdown validation for position column (B3:B1000)
	dvRange := excelize.NewDataValidation(true)
	dvRange.Sqref = "B3:B1000"
	_ = dvRange.SetDropList(positionNames)
	_ = f.AddDataValidation(sheetName, dvRange)
}

func addPartyValidation(f *excelize.File, sheetName string, parties []models.PoliticalParty) {
	// Create list of party names for validation
	partyNames := make([]string, len(parties))
	for i, p := range parties {
		partyNames[i] = p.Name
	}

	// Add dropdown validation for party column (E3:E1000)
	dvRange := excelize.NewDataValidation(true)
	dvRange.Sqref = "E3:E1000"
	_ = dvRange.SetDropList(partyNames)
	_ = f.AddDataValidation(sheetName, dvRange)
}

func addJurisdictionTypeValidation(f *excelize.File, sheetName string) {
	// Jurisdiction types
	types := []string{"national", "region", "province", "city", "barangay", "district"}

	// Add dropdown validation for jurisdiction type column (C3:C1000)
	dvRange := excelize.NewDataValidation(true)
	dvRange.Sqref = "C3:C1000"
	_ = dvRange.SetDropList(types)
	_ = f.AddDataValidation(sheetName, dvRange)
}

func createReferenceSheet[T any](f *excelize.File, sheetName string, items []T, rowFunc func(T) []string, headers []string) {
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// Headers
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9D9D9"}, Pattern: 1},
	})

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Data
	for i, item := range items {
		row := i + 2
		cols := rowFunc(item)
		for j, val := range cols {
			cell := fmt.Sprintf("%s%d", string(rune('A'+j)), row)
			_ = f.SetCellValue(sheetName, cell, val)
		}
	}

	// Auto-fit columns
	for i := range headers {
		col := string(rune('A' + i))
		_ = f.SetColWidth(sheetName, col, col, 25)
	}
}

func createInstructionsSheet(f *excelize.File) {
	sheetName := "Instructions"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	instructions := []string{
		"POLITICIAN IMPORT TEMPLATE - INSTRUCTIONS",
		"",
		"Required Fields (marked with *):",
		"  • Name: Full name of the politician",
		"  • Position: Must match exactly with a position from 'Valid Positions' sheet",
		"  • Jurisdiction Type: One of: national, region, province, city, barangay, district",
		"  • Jurisdiction Name: Name of the jurisdiction (e.g., 'Makati City', 'Metro Manila')",
		"  • Party: Must match exactly with a party from 'Valid Parties' sheet",
		"  • Term Start: Date in format YYYY-MM-DD (e.g., 2022-06-30)",
		"",
		"Optional Fields:",
		"  • Term End: Date in format YYYY-MM-DD",
		"  • Photo URL: URL to politician's photo",
		"  • Short Bio: Brief biography or description",
		"",
		"Important Notes:",
		"  1. Do not modify the header row (row 1)",
		"  2. Row 2 contains examples - you can delete it before importing",
		"  3. Use the dropdown menus for Position, Party, and Jurisdiction Type",
		"  4. Dates must be in YYYY-MM-DD format",
		"  5. Position uniqueness: Only one politician can hold a position in a jurisdiction at a time",
		"  6. If importing the same politician to the same position, it will update without creating history",
		"  7. If importing a different politician to an occupied position, the current holder will be archived",
		"",
		"For help or questions, contact your administrator.",
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})

	for i, instruction := range instructions {
		cell := fmt.Sprintf("A%d", i+1)
		_ = f.SetCellValue(sheetName, cell, instruction)
		if i == 0 {
			_ = f.SetCellStyle(sheetName, cell, cell, titleStyle)
		}
	}

	_ = f.SetColWidth(sheetName, "A", "A", 100)
}

// GenerateErrorReport generates an Excel file with import errors
func GenerateErrorReport(importLog *models.PoliticianImportLog, rows []models.ImportRow) (*excelize.File, error) {
	f := excelize.NewFile()

	sheetName := "Import Errors"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)
	_ = f.DeleteSheet("Sheet1")

	// Summary section
	_ = f.SetCellValue(sheetName, "A1", "IMPORT ERROR REPORT")
	_ = f.SetCellValue(sheetName, "A2", "Filename:")
	_ = f.SetCellValue(sheetName, "B2", importLog.Filename)
	_ = f.SetCellValue(sheetName, "A3", "Import Date:")
	_ = f.SetCellValue(sheetName, "B3", time.Now().Format("2006-01-02 15:04:05"))
	_ = f.SetCellValue(sheetName, "A4", "Total Rows:")
	_ = f.SetCellValue(sheetName, "B4", importLog.TotalRows)
	_ = f.SetCellValue(sheetName, "A5", "Successful:")
	_ = f.SetCellValue(sheetName, "B5", importLog.SuccessfulImports)
	_ = f.SetCellValue(sheetName, "A6", "Failed:")
	_ = f.SetCellValue(sheetName, "B6", importLog.FailedImports)

	// Error details headers
	headers := []string{"Row", "Field", "Error", "Value", "Suggestions"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s8", string(rune('A'+i)))
		_ = f.SetCellValue(sheetName, cell, header)
	}

	// Write errors
	row := 9
	for _, err := range importLog.ValidationErrors {
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), err.Row)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), err.Field)
		_ = f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), err.Error)
		if err.Value != nil {
			_ = f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), *err.Value)
		}
		if len(err.Suggestions) > 0 {
			_ = f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("%v", err.Suggestions))
		}
		row++
	}

	return f, nil
}
