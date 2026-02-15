package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type Response struct {
	Products []Product `json:"products"`
	Total    int64     `json:"total"`
}

type Product struct {
	Code     string   `json:"code"`
	Price    float64  `json:"price"`
	Category Category `json:"category"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductLister interface {
	List(ctx context.Context, offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error)
}

type CatalogHandler struct {
	repo ProductLister
}

func NewCatalogHandler(r ProductLister) *CatalogHandler {
	return &CatalogHandler{repo: r}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	offset, err := parseOffset(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "invalid offset", http.StatusBadRequest)
		return
	}

	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "invalid limit", http.StatusBadRequest)
		return
	}

	filters, err := parseFilters(r)
	if err != nil {
		http.Error(w, "invalid filters", http.StatusBadRequest)
		return
	}

	res, total, err := h.repo.List(r.Context(), offset, limit, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
			Category: Category{
				Code: p.Category.Code,
				Name: p.Category.Name,
			},
		}
	}

	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Products: products,
		Total:    total,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func parseOffset(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, nil
	}

	return offset, nil
}

func parseLimit(raw string) (int, error) {
	const (
		defaultLimit = 10
		minLimit     = 1
		maxLimit     = 100
	)

	if raw == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if limit < minLimit {
		return 0, errors.New("limit below minimum")
	}
	if limit > maxLimit {
		return 0, errors.New("limit above maximum")
	}

	return limit, nil
}

func parseFilters(r *http.Request) (models.ProductFilters, error) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))

	priceLessThan, err := parsePriceLessThan(r)
	if err != nil {
		return models.ProductFilters{}, err
	}

	return models.ProductFilters{
		Category:      category,
		PriceLessThan: priceLessThan,
	}, nil
}

func parsePriceLessThan(r *http.Request) (*decimal.Decimal, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("price_less_than"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("priceLessThan"))
	}
	if raw == "" {
		return nil, nil
	}

	value, err := decimal.NewFromString(raw)
	if err != nil {
		return nil, err
	}

	return &value, nil
}
