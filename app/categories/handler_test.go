package categories

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type categoriesRepoStub struct {
	listFn   func(ctx context.Context) ([]models.ProductCategory, error)
	createFn func(ctx context.Context, category models.ProductCategory) (*models.ProductCategory, error)
}

func (s categoriesRepoStub) List(ctx context.Context) ([]models.ProductCategory, error) {
	if s.listFn == nil {
		return nil, nil
	}

	return s.listFn(ctx)
}

func (s categoriesRepoStub) Create(ctx context.Context, category models.ProductCategory) (*models.ProductCategory, error) {
	if s.createFn == nil {
		return nil, nil
	}

	return s.createFn(ctx, category)
}

func TestHandleGet(t *testing.T) {
	t.Run("returns categories list", func(t *testing.T) {
		h := NewHandler(categoriesRepoStub{
			listFn: func(_ context.Context) ([]models.ProductCategory, error) {
				return []models.ProductCategory{
					{Code: "CLOTHING", Name: "Clothing"},
					{Code: "SHOES", Name: "Shoes"},
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		h.HandleGet(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{
			"categories":[
				{"code":"CLOTHING","name":"Clothing"},
				{"code":"SHOES","name":"Shoes"}
			]
		}`, rec.Body.String())
	})

	t.Run("returns 500 on repository error", func(t *testing.T) {
		h := NewHandler(categoriesRepoStub{
			listFn: func(_ context.Context) ([]models.ProductCategory, error) {
				return nil, errors.New("db error")
			},
		})

		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		rec := httptest.NewRecorder()
		h.HandleGet(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestHandlePost(t *testing.T) {
	t.Run("creates category", func(t *testing.T) {
		h := NewHandler(categoriesRepoStub{
			createFn: func(_ context.Context, category models.ProductCategory) (*models.ProductCategory, error) {
				assert.Equal(t, "ACCESSORIES_PLUS", category.Code)
				assert.Equal(t, "Accessories Plus", category.Name)

				return &models.ProductCategory{
					ID:   10,
					Code: category.Code,
					Name: category.Name,
				}, nil
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":"ACCESSORIES_PLUS","name":"Accessories Plus"}`))
		rec := httptest.NewRecorder()
		h.HandlePost(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"code":"ACCESSORIES_PLUS","name":"Accessories Plus"}`, rec.Body.String())
	})

	t.Run("returns 400 for invalid body", func(t *testing.T) {
		h := NewHandler(categoriesRepoStub{})

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{`))
		rec := httptest.NewRecorder()
		h.HandlePost(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 when required fields are empty", func(t *testing.T) {
		h := NewHandler(categoriesRepoStub{})

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":" ","name":" "}`))
		rec := httptest.NewRecorder()
		h.HandlePost(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 500 on repository error", func(t *testing.T) {
		h := NewHandler(categoriesRepoStub{
			createFn: func(_ context.Context, _ models.ProductCategory) (*models.ProductCategory, error) {
				return nil, errors.New("insert failed")
			},
		})

		req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"code":"ABC","name":"Abc"}`))
		rec := httptest.NewRecorder()
		h.HandlePost(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
