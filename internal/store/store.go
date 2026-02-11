package store

import (
	"context"
	"inventory-cli/internal/domain"
)

// ProductStore defines the interface for product storage operations.
type ProductStore interface {
	Create(ctx context.Context, product domain.Product) error
	Get(ctx context.Context, id string) (domain.Product, error)
	Update(ctx context.Context, id string, product domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter domain.ListFilter) ([]domain.Product, error)
	BulkImport(ctx context.Context, products []domain.Product) error
}
