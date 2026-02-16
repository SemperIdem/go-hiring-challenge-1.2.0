package models

import (
	"context"

	"gorm.io/gorm"
)

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) List(ctx context.Context) ([]ProductCategory, error) {
	var categories []ProductCategory

	if err := r.db.WithContext(ctx).
		Order("id ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *CategoriesRepository) Create(ctx context.Context, category ProductCategory) (*ProductCategory, error) {
	if err := r.db.WithContext(ctx).Create(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}
