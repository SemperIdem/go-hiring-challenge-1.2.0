package categories

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CategoryRepository interface {
	List(ctx context.Context) ([]models.ProductCategory, error)
	Create(ctx context.Context, category models.ProductCategory) (*models.ProductCategory, error)
}

type Handler struct {
	repo CategoryRepository
}

type listResponse struct {
	Categories []categoryResponse `json:"categories"`
}

type categoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type createCategoryRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func NewHandler(repo CategoryRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.List(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	responseCategories := make([]categoryResponse, len(categories))
	for i, category := range categories {
		responseCategories[i] = categoryResponse{
			Code: category.Code,
			Name: category.Name,
		}
	}

	api.OKResponse(w, listResponse{Categories: responseCategories})
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	var payload createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "invalid body")
		return
	}

	payload.Code = strings.TrimSpace(payload.Code)
	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Code == "" || payload.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "code and name are required")
		return
	}

	created, err := h.repo.Create(r.Context(), models.ProductCategory{
		Code: payload.Code,
		Name: payload.Name,
	})
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(categoryResponse{
		Code: created.Code,
		Name: created.Name,
	}); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
}
