package handlers

import (
	"net/http"

	"github.com/humfurie/pulpulitiko/api/internal/repository"
)

type MetricsHandler struct {
	metricsRepo *repository.MetricsRepository
}

func NewMetricsHandler(metricsRepo *repository.MetricsRepository) *MetricsHandler {
	return &MetricsHandler{metricsRepo: metricsRepo}
}

func (h *MetricsHandler) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.metricsRepo.GetDashboardMetrics(ctx)
	if err != nil {
		WriteInternalError(w, "Failed to get metrics")
		return
	}

	WriteSuccess(w, metrics)
}

func (h *MetricsHandler) GetTopArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	articles, err := h.metricsRepo.GetTopArticles(ctx, 10)
	if err != nil {
		WriteInternalError(w, "Failed to get top articles")
		return
	}

	WriteSuccess(w, articles)
}

func (h *MetricsHandler) GetCategoryMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.metricsRepo.GetCategoryMetrics(ctx)
	if err != nil {
		WriteInternalError(w, "Failed to get category metrics")
		return
	}

	WriteSuccess(w, metrics)
}

func (h *MetricsHandler) GetTagMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.metricsRepo.GetTagMetrics(ctx)
	if err != nil {
		WriteInternalError(w, "Failed to get tag metrics")
		return
	}

	WriteSuccess(w, metrics)
}
