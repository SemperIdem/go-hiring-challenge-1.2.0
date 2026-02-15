package models

// ProductCategory represents a category for products in the catalog.
type ProductCategory struct {
	ID   uint   `gorm:"primaryKey"`
	Code string `gorm:"uniqueIndex;not null"`
	Name string `gorm:"not null"`
}

func (c *ProductCategory) TableName() string {
	return "product_categories"
}
