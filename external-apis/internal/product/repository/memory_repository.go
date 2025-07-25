package repository

import (
	"errors"
	"math/big"
	"sync"

	"external-apis/internal/product/model"
	"github.com/google/uuid"
)

// ProductRepository defines the interface for product operations
type ProductRepository interface {
	GetByID(id string) (*model.Product, error)
	GetAll() ([]*model.Product, error)
	Create(product *model.Product) (*model.Product, error)
	Update(id string, product *model.Product) (*model.Product, error)
	Delete(id string) error
	ExistsByID(id string) bool
}

// MemoryProductRepository implements ProductRepository using in-memory storage
type MemoryProductRepository struct {
	products map[string]*model.Product
	mutex    sync.RWMutex
}

// NewMemoryProductRepository creates a new in-memory product repository
func NewMemoryProductRepository() *MemoryProductRepository {
	repo := &MemoryProductRepository{
		products: make(map[string]*model.Product),
	}

	// Initialize with sample data
	repo.initSampleData()

	return repo
}

// GetByID retrieves a product by ID
func (r *MemoryProductRepository) GetByID(id string) (*model.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, errors.New("product not found")
	}

	return product, nil
}

// GetAll retrieves all products
func (r *MemoryProductRepository) GetAll() ([]*model.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	products := make([]*model.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}

	return products, nil
}

// Create creates a new product
func (r *MemoryProductRepository) Create(product *model.Product) (*model.Product, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if product.ID == "" {
		product.ID = uuid.New().String()
	}

	if r.existsByIDUnsafe(product.ID) {
		return nil, errors.New("product already exists")
	}

	r.products[product.ID] = product
	return product, nil
}

// Update updates an existing product
func (r *MemoryProductRepository) Update(id string, product *model.Product) (*model.Product, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.existsByIDUnsafe(id) {
		return nil, errors.New("product not found")
	}

	product.ID = id
	r.products[id] = product
	return product, nil
}

// Delete deletes a product by ID
func (r *MemoryProductRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.existsByIDUnsafe(id) {
		return errors.New("product not found")
	}

	delete(r.products, id)
	return nil
}

// ExistsByID checks if a product exists by ID
func (r *MemoryProductRepository) ExistsByID(id string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.existsByIDUnsafe(id)
}

// existsByIDUnsafe checks if a product exists by ID (without locking)
func (r *MemoryProductRepository) existsByIDUnsafe(id string) bool {
	_, exists := r.products[id]
	return exists
}

// initSampleData initializes the repository with sample data
func (r *MemoryProductRepository) initSampleData() {
	sampleProducts := []*model.Product{
		{
			ID:          "product-789",
			Name:        "Laptop",
			Description: "High-performance laptop for professional use",
			Price:       big.NewRat(99900, 100), // 999.00
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-001",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse with precision tracking",
			Price:       big.NewRat(2999, 100), // 29.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-002",
			Name:        "Mechanical Keyboard",
			Description: "RGB mechanical keyboard with Cherry MX switches",
			Price:       big.NewRat(12999, 100), // 129.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-003",
			Name:        "4K Monitor",
			Description: "27-inch 4K UHD monitor with HDR support",
			Price:       big.NewRat(39999, 100), // 399.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-004",
			Name:        "USB-C Hub",
			Description: "Multi-port USB-C hub with HDMI and Ethernet",
			Price:       big.NewRat(7999, 100), // 79.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-005",
			Name:        "Bluetooth Headphones",
			Description: "Noise-cancelling wireless headphones",
			Price:       big.NewRat(19999, 100), // 199.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-006",
			Name:        "Smartphone",
			Description: "Latest smartphone with advanced camera",
			Price:       big.NewRat(79999, 100), // 799.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-007",
			Name:        "Tablet",
			Description: "10-inch tablet with stylus support",
			Price:       big.NewRat(49999, 100), // 499.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-008",
			Name:        "Smartwatch",
			Description: "Fitness tracking smartwatch with GPS",
			Price:       big.NewRat(29999, 100), // 299.99
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-inactive",
			Name:        "Discontinued Product",
			Description: "This product is no longer available",
			Price:       big.NewRat(9999, 100), // 99.99
			Category:    "Electronics",
			Active:      false,
		},
	}

	for _, product := range sampleProducts {
		r.products[product.ID] = product
	}
}
