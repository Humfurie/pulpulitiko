package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/humfurie/pulpulitiko/api/internal/models"
	"github.com/humfurie/pulpulitiko/api/internal/services"
	"github.com/humfurie/pulpulitiko/api/pkg/excel"
)

type ImportHandler struct {
	importService *services.ImportService
}

func NewImportHandler(importService *services.ImportService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// POST /api/admin/import/politicians/validate - Validate Excel file without importing
func (h *ImportHandler) ValidatePoliticianImport(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		WriteBadRequest(w, "failed to parse form data")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteBadRequest(w, "file is required")
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		WriteInternalError(w, "failed to read file")
		return
	}

	// Parse Excel file
	rows, err := excel.ParseImportFile(fileData)
	if err != nil {
		WriteBadRequest(w, fmt.Sprintf("failed to parse Excel file: %s", err.Error()))
		return
	}

	// Validate rows
	result, err := h.importService.ValidateImport(r.Context(), rows)
	if err != nil {
		WriteInternalError(w, fmt.Sprintf("validation failed: %s", err.Error()))
		return
	}

	// Return validation result
	WriteSuccess(w, map[string]interface{}{
		"filename":       header.Filename,
		"total_rows":     result.TotalRows,
		"valid_rows":     result.ValidRows,
		"invalid_rows":   result.InvalidRows,
		"errors":         result.Errors,
		"validated_rows": result.ValidatedRows,
	})
}

// POST /api/admin/import/politicians - Import politicians from Excel
func (h *ImportHandler) ImportPoliticians(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		WriteBadRequest(w, "failed to parse form data")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		WriteBadRequest(w, "file is required")
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		WriteInternalError(w, "failed to read file")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := GetUserIDFromRequest(r)

	// Optional: Get election ID from form (for linking import to election)
	var electionID *uuid.UUID
	if electionIDStr := r.FormValue("election_id"); electionIDStr != "" {
		parsedID, err := uuid.Parse(electionIDStr)
		if err == nil {
			electionID = &parsedID
		}
	}

	// Create import request
	importReq := &models.ProcessImportRequest{
		FileData:   fileData,
		Filename:   header.Filename,
		ElectionID: electionID,
	}

	// Start import process (async)
	importLog, err := h.importService.StartImport(r.Context(), importReq, userID)
	if err != nil {
		WriteInternalError(w, fmt.Sprintf("failed to start import: %s", err.Error()))
		return
	}

	WriteCreated(w, importLog)
}

// GET /api/admin/import/politicians/logs - List all import logs
func (h *ImportHandler) ListImportLogs(w http.ResponseWriter, r *http.Request) {
	page, perPage := GetPaginationParams(r)

	logs, err := h.importService.ListImportLogs(r.Context(), page, perPage)
	if err != nil {
		WriteInternalError(w, "failed to fetch import logs")
		return
	}

	WriteSuccess(w, logs)
}

// GET /api/admin/import/politicians/logs/:id - Get single import log details
func (h *ImportHandler) GetImportLog(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid import log ID")
		return
	}

	log, err := h.importService.GetImportLog(r.Context(), id)
	if err != nil {
		WriteInternalError(w, "failed to fetch import log")
		return
	}

	if log == nil {
		WriteNotFound(w, "import log not found")
		return
	}

	WriteSuccess(w, log)
}

// GET /api/admin/import/politicians/template - Download Excel template
func (h *ImportHandler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	// Generate template file
	templateFile, err := h.importService.GenerateTemplate(r.Context())
	if err != nil {
		WriteInternalError(w, fmt.Sprintf("failed to generate template: %s", err.Error()))
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=politician_import_template.xlsx")

	// Write file to response
	if err := templateFile.Write(w); err != nil {
		WriteInternalError(w, "failed to write file")
		return
	}
}

// GET /api/admin/import/politicians/logs/:id/errors - Download error report for failed import
func (h *ImportHandler) DownloadErrorReport(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteBadRequest(w, "invalid import log ID")
		return
	}

	// Generate error report
	errorFile, err := h.importService.GenerateErrorReport(r.Context(), id)
	if err != nil {
		WriteInternalError(w, fmt.Sprintf("failed to generate error report: %s", err.Error()))
		return
	}

	if errorFile == nil {
		WriteNotFound(w, "error report not available")
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=import_errors_%s.xlsx", id.String()))

	// Write file to response
	if err := errorFile.Write(w); err != nil {
		WriteInternalError(w, "failed to write file")
		return
	}
}

// Helper to get user ID from request context
func GetUserIDFromRequest(r *http.Request) *uuid.UUID {
	// This should be set by the auth middleware
	// Implementation depends on how you store user info in context
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		return nil
	}
	return &userID
}
