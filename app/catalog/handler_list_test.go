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

type catalogListRepoStub struct {
	listFn      func(ctx context.Context, offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error)
	getByCodeFn func(ctx context.Context, code string) (*models.Product, error)
}

func (s catalogListRepoStub) List(ctx context.Context, offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error) {
	if s.listFn == nil {
		return nil, 0, nil
	}

	return s.listFn(ctx, offset, limit, filters)
}

func (s catalogListRepoStub) GetByCode(ctx context.Context, code string) (*models.Product, error) {
	if s.getByCodeFn == nil {
		return nil, nil
	}

	return s.getByCodeFn(ctx, code)
}

func TestCatalogHandleGet(t *testing.T) {
	t.Run("uses default pagination and returns total with products", func(t *testing.T) {
		price := decimal.RequireFromString("10.99")
		handler := NewCatalogHandler(catalogListRepoStub{
			listFn: func(_ context.Context, offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error) {
				assert.Equal(t, 0, offset)
				assert.Equal(t, 10, limit)
				assert.Equal(t, "", filters.Category)
				assert.Nil(t, filters.PriceLessThan)

				return []models.Product{
					{
						Code:  "PROD001",
						Price: price,
						Category: models.ProductCategory{
							Code: "CLOTHING",
							Name: "Clothing",
						},
					},
				}, 8, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"products":[
				{"code":"PROD001","price":10.99,"category":{"code":"CLOTHING","name":"Clothing"}}
			],
			"total":8
		}`, rec.Body.String())
	})

	t.Run("applies pagination and filters", func(t *testing.T) {
		price := decimal.RequireFromString("12.49")
		expectedLessThan := decimal.RequireFromString("15.00")
		handler := NewCatalogHandler(catalogListRepoStub{
			listFn: func(_ context.Context, offset, limit int, filters models.ProductFilters) ([]models.Product, int64, error) {
				assert.Equal(t, 2, offset)
				assert.Equal(t, 5, limit)
				assert.Equal(t, "Shoes", filters.Category)
				assert.NotNil(t, filters.PriceLessThan)
				assert.True(t, filters.PriceLessThan.Equal(expectedLessThan))

				return []models.Product{
					{
						Code:  "PROD002",
						Price: price,
						Category: models.ProductCategory{
							Code: "SHOES",
							Name: "Shoes",
						},
					},
				}, 2, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=2&limit=5&category=Shoes&price_less_than=15.00", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{
			"products":[
				{"code":"PROD002","price":12.49,"category":{"code":"SHOES","name":"Shoes"}}
			],
			"total":2
		}`, rec.Body.String())
	})

	t.Run("accepts camelCase priceLessThan filter", func(t *testing.T) {
		expectedLessThan := decimal.RequireFromString("20.00")
		handler := NewCatalogHandler(catalogListRepoStub{
			listFn: func(_ context.Context, _ int, _ int, filters models.ProductFilters) ([]models.Product, int64, error) {
				assert.NotNil(t, filters.PriceLessThan)
				assert.True(t, filters.PriceLessThan.Equal(expectedLessThan))
				return []models.Product{}, 0, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=20.00", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("returns 400 for invalid offset", func(t *testing.T) {
		handler := NewCatalogHandler(catalogListRepoStub{})
		req := httptest.NewRequest(http.MethodGet, "/catalog?offset=abc", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"invalid offset"}`, rec.Body.String())
	})

	t.Run("returns 400 for invalid limit", func(t *testing.T) {
		handler := NewCatalogHandler(catalogListRepoStub{})
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=101", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"invalid limit"}`, rec.Body.String())
	})

	t.Run("returns 400 for limit below minimum", func(t *testing.T) {
		handler := NewCatalogHandler(catalogListRepoStub{})
		req := httptest.NewRequest(http.MethodGet, "/catalog?limit=0", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"invalid limit"}`, rec.Body.String())
	})

	t.Run("returns 400 for invalid filters", func(t *testing.T) {
		handler := NewCatalogHandler(catalogListRepoStub{})
		req := httptest.NewRequest(http.MethodGet, "/catalog?price_less_than=abc", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.JSONEq(t, `{"error":"invalid filters"}`, rec.Body.String())
	})

	t.Run("returns 500 when repository list fails", func(t *testing.T) {
		handler := NewCatalogHandler(catalogListRepoStub{
			listFn: func(_ context.Context, _ int, _ int, _ models.ProductFilters) ([]models.Product, int64, error) {
				return nil, 0, errors.New("db unavailable")
			},
		})
		req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
		rec := httptest.NewRecorder()

		handler.HandleGet(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error":"db unavailable"}`, rec.Body.String())
	})
}
