package domain

import (
	"fmt"
)

// Product represents a product in the inventory.
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Category string  `json:"category"`
}

// ListFilter defines criteria for filtering products.
type ListFilter struct {
	Category *string  // Optional: Filter by category
	MinPrice *float64 // Optional: Minimum price
	MaxPrice *float64 // Optional: Maximum price
}

// Custom Error Types

// ProductNotFoundError is returned when a product with the given ID is not found.
type ProductNotFoundError struct {
	ID string
}

func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product with ID %s not found", e.ID)
}

// InvalidProductError is returned when product validation fails.
type InvalidProductError struct {
	Details string
}

func (e *InvalidProductError) Error() string {
	return fmt.Sprintf("invalid product: %s", e.Details)
}

// DuplicateProductError is returned when attempting to create a product with an existing ID.
type DuplicateProductError struct {
	ID string
}

func (e *DuplicateProductError) Error() string {
	return fmt.Sprintf("product with ID %s already exists", e.ID)
}
