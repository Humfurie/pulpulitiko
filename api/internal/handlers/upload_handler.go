package handlers

import (
	"net/http"

	"github.com/humfurie/pulpulitiko/api/internal/services"
	"github.com/humfurie/pulpulitiko/api/pkg/storage"
)

type UploadHandler struct {
	uploadService *services.UploadService
}

func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// POST /api/admin/upload
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, storage.GetMaxFileSize()+1024)

	if err := r.ParseMultipartForm(storage.GetMaxFileSize()); err != nil {
		WriteBadRequest(w, "file too large or invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		WriteBadRequest(w, "file is required")
		return
	}
	defer file.Close()

	result, err := h.uploadService.UploadFile(r.Context(), file, header)
	if err != nil {
		WriteBadRequest(w, err.Error())
		return
	}

	WriteSuccess(w, result)
}
