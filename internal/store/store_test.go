package store

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/rohitaj002/product-inventory-CLI/internal/domain"
)

func TestInMemoryStore_CRUD(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	p := domain.Product{
		ID:       "1",
		Name:     "Test Product",
		Price:    10.0,
		Quantity: 5,
		Category: "Test",
	}

	// Create
	err := store.Create(ctx, p)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get
	got, err := store.Get(ctx, "1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("Expected name %s, got %s", p.Name, got.Name)
	}

	// Update
	p.Price = 20.0
	err = store.Update(ctx, "1", p)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _ = store.Get(ctx, "1")
	if got.Price != 20.0 {
		t.Errorf("Expected price 20.0, got %f", got.Price)
	}

	// List
	list, err := store.List(ctx, domain.ListFilter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 product, got %d", len(list))
	}

	// Delete
	err = store.Delete(ctx, "1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get(ctx, "1")
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestInMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent Create
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			p := domain.Product{ID: id, Name: "Product " + id, Price: 10.0}
			store.Create(ctx, p)
		}(string(rune(i))) // simple id generation
	}
	wg.Wait()

	list, _ := store.List(ctx, domain.ListFilter{})
	if len(list) != 100 {
		t.Errorf("Expected 100 products, got %d", len(list))
	}
}

func TestJSONFileStore_Persistence(t *testing.T) {
	tmpFile := "test_store.json"
	defer os.Remove(tmpFile)

	store1, _ := NewJSONFileStore(tmpFile)
	ctx := context.Background()

	p := domain.Product{ID: "1", Name: "Persistent Product", Price: 100}
	store1.Create(ctx, p)

	// Open new store from same file
	store2, err := NewJSONFileStore(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open second store: %v", err)
	}

	got, err := store2.Get(ctx, "1")
	if err != nil {
		t.Fatalf("Failed to get product from reloaded store: %v", err)
	}
	if got.Name != p.Name {
		t.Errorf("Expected name %s, got %s", p.Name, got.Name)
	}
}
