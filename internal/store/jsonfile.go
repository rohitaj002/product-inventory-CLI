package store

import (
	"context"
	"encoding/json"
	"fmt"
	"inventory-cli/internal/domain"
	"os"
	"sync"
)

// JSONFileStore extends InMemoryStore with JSON file persistence.
type JSONFileStore struct {
	*InMemoryStore
	filePath string
	fileMu   sync.Mutex
}

// NewJSONFileStore creates a new JSONFileStore and loads data if file exists.
func NewJSONFileStore(filePath string) (*JSONFileStore, error) {
	store := &JSONFileStore{
		InMemoryStore: NewInMemoryStore(),
		filePath:      filePath,
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *JSONFileStore) load() error {
	s.fileMu.Lock()
	defer s.fileMu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return nil // New file, empty store
	}
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var products map[string]domain.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return err
	}

	// Lock the memory store to populate it
	s.mu.Lock()
	if products == nil {
		s.products = make(map[string]domain.Product)
	} else {
		s.products = products
	}
	s.mu.Unlock()

	return nil
}

func (s *JSONFileStore) save() error {
	s.fileMu.Lock()
	defer s.fileMu.Unlock()

	s.mu.RLock()
	data, err := json.MarshalIndent(s.products, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// Override modifying methods to trigger save

func (s *JSONFileStore) Create(ctx context.Context, product domain.Product) error {
	if err := s.InMemoryStore.Create(ctx, product); err != nil {
		return err
	}
	return s.save()
}

func (s *JSONFileStore) Update(ctx context.Context, id string, product domain.Product) error {
	if err := s.InMemoryStore.Update(ctx, id, product); err != nil {
		return err
	}
	return s.save()
}

func (s *JSONFileStore) Delete(ctx context.Context, id string) error {
	if err := s.InMemoryStore.Delete(ctx, id); err != nil {
		return err
	}
	return s.save()
}

func (s *JSONFileStore) BulkImport(ctx context.Context, products []domain.Product) error {
	importErr := s.InMemoryStore.BulkImport(ctx, products)
	if err := s.save(); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}
	return importErr
}
