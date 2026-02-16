package models

import (
	"context"
	"errors"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

type ProductFilters struct {
	Category      string
	PriceLessThan *decimal.Decimal
}

var ErrProductNotFound = errors.New("product not found")

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

func (r *ProductsRepository) List(ctx context.Context, offset, limit int, filters ProductFilters) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.WithContext(ctx).Model(&Product{})

	if filters.Category != "" {
		categoryFilter := strings.TrimSpace(filters.Category)
		query = query.
			Joins("JOIN product_categories ON product_categories.id = products.category_id").
			Where(
				"product_categories.code = ? OR product_categories.name ILIKE ?",
				strings.ToUpper(categoryFilter),
				categoryFilter,
			)
	}

	if filters.PriceLessThan != nil {
		query = query.Where("products.price < ?", filters.PriceLessThan.String())
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Variants").
		Preload("Category").
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetByCode(ctx context.Context, code string) (*Product, error) {
	var product Product

	if err := r.db.WithContext(ctx).
		Preload("Variants").
		Preload("Category").
		Where("code = ?", code).
		First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}
