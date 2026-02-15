package catalog

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type productDetailsRepoStub struct {
	getByCodeFn func(ctx context.Context, code string) (*models.Product, error)
}

func (s productDetailsRepoStub) List(ctx context.Context, offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error) {
	return nil, 0, nil
}

func (s productDetailsRepoStub) GetByCode(ctx context.Context, code string) (*models.Product, error) {
	if s.getByCodeFn == nil {
		return nil, nil
	}

	return s.getByCodeFn(ctx, code)
}

func TestCatalogHandleGetByCode(t *testing.T) {
	t.Run("returns product details with category and variant price fallback", func(t *testing.T) {
		productPrice := decimal.RequireFromString("10.99")
		variantPrice := decimal.RequireFromString("11.99")

		h := NewCatalogHandler(productDetailsRepoStub{
			getByCodeFn: func(_ context.Context, code string) (*models.Product, error) {
				assert.Equal(t, "PROD001", code)
				return &models.Product{
					Code:  "PROD001",
					Price: productPrice,
					Category: models.ProductCategory{
						Code: "CLOTHING",
						Name: "Clothing",
					},
					Variants: []models.Variant{
						{
							Name:  "Variant A",
							SKU:   "SKU001A",
							Price: &variantPrice,
						},
						{
							Name: "Variant B",
							SKU:  "SKU001B",
						},
					},
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		rec := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /catalog/{code}", h.HandleGetByCode)

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"code":"PROD001",
			"price":10.99,
			"category":{"code":"CLOTHING","name":"Clothing"},
			"variants":[
				{"name":"Variant A","sku":"SKU001A","price":11.99},
				{"name":"Variant B","sku":"SKU001B","price":10.99}
			]
		}`, rec.Body.String())
	})

	t.Run("returns 404 when product does not exist", func(t *testing.T) {
		h := NewCatalogHandler(productDetailsRepoStub{
			getByCodeFn: func(_ context.Context, _ string) (*models.Product, error) {
				return nil, models.ErrProductNotFound
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/catalog/UNKNOWN", nil)
		rec := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /catalog/{code}", h.HandleGetByCode)

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		h := NewCatalogHandler(productDetailsRepoStub{
			getByCodeFn: func(_ context.Context, _ string) (*models.Product, error) {
				return nil, errors.New("db unavailable")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/catalog/PROD001", nil)
		rec := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.HandleFunc("GET /catalog/{code}", h.HandleGetByCode)

		mux.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
