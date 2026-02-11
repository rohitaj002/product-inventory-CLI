package store

import (
	"context"
	"errors"
	"fmt"
	"inventory-cli/internal/domain"
	"log/slog"
	"sync"
)

type InMemoryStore struct {
	mu       sync.RWMutex
	products map[string]domain.Product
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		products: make(map[string]domain.Product),
	}
}

func (s *InMemoryStore) Create(ctx context.Context, product domain.Product) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[product.ID]; exists {
		slog.Warn("Attempted to create duplicate product", "id", product.ID)
		return &domain.DuplicateProductError{ID: product.ID}
	}

	s.products[product.ID] = product
	slog.Info("Product created", "id", product.ID, "name", product.Name)
	return nil
}

func (s *InMemoryStore) Get(ctx context.Context, id string) (domain.Product, error) {
	select {
	case <-ctx.Done():
		return domain.Product{}, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[id]
	if !exists {
		return domain.Product{}, &domain.ProductNotFoundError{ID: id}
	}

	return product, nil
}

func (s *InMemoryStore) Update(ctx context.Context, id string, product domain.Product) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		slog.Warn("Attempted to update non-existent product", "id", id)
		return &domain.ProductNotFoundError{ID: id}
	}

	// Ensure ID matches
	if product.ID != id {
		return errors.New("product ID mismatch")
	}

	s.products[id] = product
	slog.Info("Product updated", "id", id)
	return nil
}

func (s *InMemoryStore) Delete(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		slog.Warn("Attempted to delete non-existent product", "id", id)
		return &domain.ProductNotFoundError{ID: id}
	}

	delete(s.products, id)
	slog.Info("Product deleted", "id", id)
	return nil
}

func (s *InMemoryStore) List(ctx context.Context, filter domain.ListFilter) ([]domain.Product, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Product
	for _, p := range s.products {
		if filter.Category != nil && p.Category != *filter.Category {
			continue
		}
		if filter.MinPrice != nil && p.Price < *filter.MinPrice {
			continue
		}
		if filter.MaxPrice != nil && p.Price > *filter.MaxPrice {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

// BulkImport implements concurrent import as per Task 6
func (s *InMemoryStore) BulkImport(ctx context.Context, products []domain.Product) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	slog.Info("Starting bulk import", "count", len(products))

	// Create a channel for jobs
	jobs := make(chan domain.Product, len(products))
	results := make(chan error, len(products))

	// Worker pool size
	numWorkers := 10
	if len(products) < numWorkers {
		numWorkers = len(products)
	}

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				// Check context cancellation inside worker
				select {
				case <-ctx.Done():
					results <- ctx.Err()
					return
				default:
				}

				// validation or other processing logic could go here
				// For now, we reuse Create which handles locking
				// Note: Create handles its own context check, but we pass ctx
				err := s.Create(ctx, p)

				// If Create fails (e.g. duplicate), we might want to return error or ignore?
				// Task 6 says: "Handles partial failures gracefully" and "Collects and aggregates results"
				// For this implementation, we'll pipe errors to results
				results <- err
			}
		}()
	}

	// Send jobs
	for _, p := range products {
		jobs <- p
	}
	close(jobs)

	// Wait for workers in a separate goroutine to close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregation
	var errs []error
	for err := range results {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		slog.Warn("Bulk import completed with errors", "error_count", len(errs))
		return fmt.Errorf("bulk import encountered %d errors: %v", len(errs), errs[0])
	}

	slog.Info("Bulk import completed successfully")
	return nil
}
