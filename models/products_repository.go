package models

import (
	"context"

	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

func (r *ProductsRepository) List(ctx context.Context) ([]Product, error) {
	var products []Product

	if err := r.db.WithContext(ctx).Preload("Variants").Preload("Category").Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}
