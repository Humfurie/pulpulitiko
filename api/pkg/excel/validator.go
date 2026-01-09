package excel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type Validator struct {
	positionRepo *repository.PoliticalPartyRepository // Has methods for positions
	partyRepo    *repository.PoliticalPartyRepository
	locationRepo *repository.LocationRepository
}

func NewValidator(
	positionRepo *repository.PoliticalPartyRepository,
	partyRepo *repository.PoliticalPartyRepository,
	locationRepo *repository.LocationRepository,
) *Validator {
	return &Validator{
		positionRepo: positionRepo,
		partyRepo:    partyRepo,
		locationRepo: locationRepo,
	}
}

// ValidateImportRows validates all import rows and returns validated rows with errors
func (v *Validator) ValidateImportRows(ctx context.Context, rows []models.ImportRow) (*models.ImportValidationResult, error) {
	result := &models.ImportValidationResult{
		TotalRows: len(rows),
		Errors:    []models.ValidationError{},
	}

	// Pre-load reference data for validation
	positions, err := v.loadPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load positions: %w", err)
	}

	parties, err := v.loadParties(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load parties: %w", err)
	}

	// Validate each row
	for _, row := range rows {
		validatedRow := v.validateRow(ctx, row, positions, parties)
		result.ValidatedRows = append(result.ValidatedRows, validatedRow)

		if !validatedRow.IsValid {
			result.InvalidRows++
			result.Errors = append(result.Errors, validatedRow.Errors...)
		} else {
			result.ValidRows++
		}
	}

	return result, nil
}

// validateRow validates a single import row
func (v *Validator) validateRow(
	ctx context.Context,
	row models.ImportRow,
	positions map[string]models.GovernmentPosition,
	parties map[string]models.PoliticalParty,
) models.ValidatedImportRow {
	validated := models.ValidatedImportRow{
		RowNumber: row.RowNumber,
		Name:      row.Name,
		IsValid:   true,
		Errors:    []models.ValidationError{},
	}

	// Validate Name (required)
	if strings.TrimSpace(row.Name) == "" {
		addError(&validated, "name", "Name is required", &row.Name)
	}

	// Validate Position (required, must exist)
	if strings.TrimSpace(row.Position) == "" {
		addError(&validated, "position", "Position is required", &row.Position)
	} else {
		position, exists := positions[strings.ToLower(row.Position)]
		if !exists {
			// Find suggestions
			suggestions := findSimilarStrings(row.Position, getPositionNames(positions), 3)
			addErrorWithSuggestions(&validated, "position",
				fmt.Sprintf("Position '%s' not found", row.Position),
				&row.Position, suggestions)
		} else {
			validated.PositionID = position.ID
			validated.PositionName = position.Name
		}
	}

	// Validate Jurisdiction Type (required)
	validTypes := []string{"national", "region", "province", "city", "barangay", "district"}
	if strings.TrimSpace(row.JurisdictionType) == "" {
		addError(&validated, "jurisdiction_type", "Jurisdiction type is required", &row.JurisdictionType)
	} else if !contains(validTypes, row.JurisdictionType) {
		addErrorWithSuggestions(&validated, "jurisdiction_type",
			fmt.Sprintf("Invalid jurisdiction type '%s'", row.JurisdictionType),
			&row.JurisdictionType, validTypes)
	} else {
		validated.JurisdictionType = row.JurisdictionType

		// Handle national jurisdiction
		if row.JurisdictionType == "national" {
			validated.IsNational = true
		} else {
			// Validate Jurisdiction Name (required for non-national)
			if strings.TrimSpace(row.JurisdictionName) == "" {
				addError(&validated, "jurisdiction_name", "Jurisdiction name is required for non-national positions", &row.JurisdictionName)
			} else {
				// Lookup jurisdiction ID based on type
				jurisdictionID, err := v.lookupJurisdiction(ctx, row.JurisdictionType, row.JurisdictionName)
				if err != nil {
					addError(&validated, "jurisdiction_name",
						fmt.Sprintf("Jurisdiction '%s' not found for type '%s'", row.JurisdictionName, row.JurisdictionType),
						&row.JurisdictionName)
				} else {
					// Set the appropriate jurisdiction ID
					switch row.JurisdictionType {
					case "region":
						validated.RegionID = &jurisdictionID
					case "province":
						validated.ProvinceID = &jurisdictionID
					case "city":
						validated.CityID = &jurisdictionID
					case "barangay":
						validated.BarangayID = &jurisdictionID
					case "district":
						validated.DistrictID = &jurisdictionID
					}
				}
			}
		}
	}

	// Validate Party (required, must exist)
	if strings.TrimSpace(row.Party) == "" {
		addError(&validated, "party", "Party is required", &row.Party)
	} else {
		party, exists := parties[strings.ToLower(row.Party)]
		if !exists {
			suggestions := findSimilarStrings(row.Party, getPartyNames(parties), 3)
			addErrorWithSuggestions(&validated, "party",
				fmt.Sprintf("Party '%s' not found", row.Party),
				&row.Party, suggestions)
		} else {
			validated.PartyID = &party.ID
			partyName := party.Name
			validated.PartyName = &partyName
		}
	}

	// Validate Term Start (required, valid date)
	if strings.TrimSpace(row.TermStart) == "" {
		addError(&validated, "term_start", "Term start date is required", &row.TermStart)
	} else {
		termStart, err := time.Parse("2006-01-02", row.TermStart)
		if err != nil {
			addError(&validated, "term_start",
				fmt.Sprintf("Invalid date format '%s'. Expected YYYY-MM-DD", row.TermStart),
				&row.TermStart)
		} else {
			validated.TermStart = termStart
		}
	}

	// Validate Term End (optional, valid date, must be after term start)
	if row.TermEnd != nil && strings.TrimSpace(*row.TermEnd) != "" {
		termEnd, err := time.Parse("2006-01-02", *row.TermEnd)
		if err != nil {
			addError(&validated, "term_end",
				fmt.Sprintf("Invalid date format '%s'. Expected YYYY-MM-DD", *row.TermEnd),
				row.TermEnd)
		} else {
			validated.TermEnd = &termEnd
			// Check if term_end is after term_start
			if !validated.TermStart.IsZero() && termEnd.Before(validated.TermStart) {
				addError(&validated, "term_end", "Term end must be after term start", row.TermEnd)
			}
		}
	}

	// Optional fields (no validation needed, just copy)
	validated.PhotoURL = row.PhotoURL
	validated.ShortBio = row.ShortBio

	// Validate Birth Date (optional, valid date)
	if row.BirthDate != nil && strings.TrimSpace(*row.BirthDate) != "" {
		birthDate, err := time.Parse("2006-01-02", *row.BirthDate)
		if err != nil {
			addError(&validated, "birth_date",
				fmt.Sprintf("Invalid date format '%s'. Expected YYYY-MM-DD", *row.BirthDate),
				row.BirthDate)
		} else {
			validated.BirthDate = &birthDate
		}
	}

	return validated
}

// Helper functions

func addError(v *models.ValidatedImportRow, field, message string, value *string) {
	v.IsValid = false
	v.Errors = append(v.Errors, models.ValidationError{
		Row:   v.RowNumber,
		Field: field,
		Error: message,
		Value: value,
	})
}

func addErrorWithSuggestions(v *models.ValidatedImportRow, field, message string, value *string, suggestions []string) {
	v.IsValid = false
	v.Errors = append(v.Errors, models.ValidationError{
		Row:         v.RowNumber,
		Field:       field,
		Error:       message,
		Value:       value,
		Suggestions: suggestions,
	})
}

// loadPositions loads all government positions into a map for quick lookup
func (v *Validator) loadPositions(ctx context.Context) (map[string]models.GovernmentPosition, error) {
	// This would call the repository to get all positions
	// For now, returning empty map - implement when repository method is available
	positions := make(map[string]models.GovernmentPosition)

	// TODO: Implement when GetAllPositions method is available
	// positionList, err := v.positionRepo.GetAllPositions(ctx)
	// if err != nil {
	//     return nil, err
	// }
	// for _, p := range positionList {
	//     positions[strings.ToLower(p.Name)] = p
	// }

	return positions, nil
}

// loadParties loads all political parties into a map for quick lookup
func (v *Validator) loadParties(ctx context.Context) (map[string]models.PoliticalParty, error) {
	// This would call the repository to get all parties
	parties := make(map[string]models.PoliticalParty)

	// TODO: Implement when GetAllParties method is available
	// partyList, err := v.partyRepo.GetAllParties(ctx)
	// if err != nil {
	//     return nil, err
	// }
	// for _, p := range partyList {
	//     parties[strings.ToLower(p.Name)] = p
	// }

	return parties, nil
}

// lookupJurisdiction finds a jurisdiction ID by type and name
func (v *Validator) lookupJurisdiction(ctx context.Context, jurisdictionType, name string) (uuid.UUID, error) {
	// This would call the location repository based on type
	// TODO: Implement when location repository methods are available

	switch jurisdictionType {
	case "region":
		// return v.locationRepo.GetRegionByName(ctx, name)
	case "province":
		// return v.locationRepo.GetProvinceByName(ctx, name)
	case "city":
		// return v.locationRepo.GetCityByName(ctx, name)
	case "barangay":
		// return v.locationRepo.GetBarangayByName(ctx, name)
	case "district":
		// return v.locationRepo.GetDistrictByName(ctx, name)
	}

	return uuid.Nil, fmt.Errorf("jurisdiction lookup not implemented")
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func getPositionNames(positions map[string]models.GovernmentPosition) []string {
	names := make([]string, 0, len(positions))
	for _, p := range positions {
		names = append(names, p.Name)
	}
	return names
}

func getPartyNames(parties map[string]models.PoliticalParty) []string {
	names := make([]string, 0, len(parties))
	for _, p := range parties {
		names = append(names, p.Name)
	}
	return names
}

// findSimilarStrings finds similar strings using simple string distance
func findSimilarStrings(target string, candidates []string, maxResults int) []string {
	type match struct {
		str      string
		distance int
	}

	var matches []match
	target = strings.ToLower(target)

	for _, candidate := range candidates {
		candidateLower := strings.ToLower(candidate)

		// Simple similarity check: contains or starts with
		if strings.Contains(candidateLower, target) || strings.HasPrefix(candidateLower, target) {
			matches = append(matches, match{candidate, 0})
		} else if strings.Contains(target, candidateLower) || strings.HasPrefix(target, candidateLower) {
			matches = append(matches, match{candidate, 1})
		}
	}

	// Return top matches
	results := make([]string, 0, maxResults)
	for i := 0; i < len(matches) && i < maxResults; i++ {
		results = append(results, matches[i].str)
	}

	return results
}
