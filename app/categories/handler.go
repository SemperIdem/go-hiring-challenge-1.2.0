package categories

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseCategories := make([]categoryResponse, len(categories))
	for i, category := range categories {
		responseCategories[i] = categoryResponse{
			Code: category.Code,
			Name: category.Name,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(listResponse{Categories: responseCategories}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	var payload createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	payload.Code = strings.TrimSpace(payload.Code)
	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Code == "" || payload.Name == "" {
		http.Error(w, "code and name are required", http.StatusBadRequest)
		return
	}

	created, err := h.repo.Create(r.Context(), models.ProductCategory{
		Code: payload.Code,
		Name: payload.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(categoryResponse{
		Code: created.Code,
		Name: created.Name,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
