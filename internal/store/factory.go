package store

import (
	"fmt"
)

// StoreType defines the type of store to create.
type StoreType string

const (
	Memory   StoreType = "memory"
	JSONFile StoreType = "json"
)

// NewStoreFactory creates a ProductStore based on the type.
func NewStoreFactory(storeType StoreType, connectionString string) (ProductStore, error) {
	switch storeType {
	case Memory:
		return NewInMemoryStore(), nil
	case JSONFile:
		return NewJSONFileStore(connectionString)
	default:
		return nil, fmt.Errorf("unsupported store type: %s", storeType)
	}
}
