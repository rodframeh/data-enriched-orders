package repository

import (
	"errors"
	"sync"

	"external-apis/internal/customer/model"
	"github.com/google/uuid"
)

// CustomerRepository defines the interface for customer operations
type CustomerRepository interface {
	GetByID(id string) (*model.Customer, error)
	GetAll() ([]*model.Customer, error)
	Create(customer *model.Customer) (*model.Customer, error)
	Update(id string, customer *model.Customer) (*model.Customer, error)
	Delete(id string) error
	ExistsByID(id string) bool
	GetByEmail(email string) (*model.Customer, error)
}

// MemoryCustomerRepository implements CustomerRepository using in-memory storage
type MemoryCustomerRepository struct {
	customers map[string]*model.Customer
	mutex     sync.RWMutex
}

// NewMemoryCustomerRepository creates a new in-memory customer repository
func NewMemoryCustomerRepository() *MemoryCustomerRepository {
	repo := &MemoryCustomerRepository{
		customers: make(map[string]*model.Customer),
	}

	// Initialize with sample data
	repo.initSampleData()

	return repo
}

// GetByID retrieves a customer by ID
func (r *MemoryCustomerRepository) GetByID(id string) (*model.Customer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	customer, exists := r.customers[id]
	if !exists {
		return nil, errors.New("customer not found")
	}

	return customer, nil
}

// GetAll retrieves all customers
func (r *MemoryCustomerRepository) GetAll() ([]*model.Customer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	customers := make([]*model.Customer, 0, len(r.customers))
	for _, customer := range r.customers {
		customers = append(customers, customer)
	}

	return customers, nil
}

// Create creates a new customer
func (r *MemoryCustomerRepository) Create(customer *model.Customer) (*model.Customer, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if customer.ID == "" {
		customer.ID = uuid.New().String()
	}

	if r.existsByIDUnsafe(customer.ID) {
		return nil, errors.New("customer already exists")
	}

	// Check for duplicate email
	if r.existsByEmailUnsafe(customer.Email) {
		return nil, errors.New("customer with this email already exists")
	}

	r.customers[customer.ID] = customer
	return customer, nil
}

// Update updates an existing customer
func (r *MemoryCustomerRepository) Update(id string, customer *model.Customer) (*model.Customer, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.existsByIDUnsafe(id) {
		return nil, errors.New("customer not found")
	}

	// Check for duplicate email (excluding current customer)
	if existing := r.getByEmailUnsafe(customer.Email); existing != nil && existing.ID != id {
		return nil, errors.New("customer with this email already exists")
	}

	customer.ID = id
	r.customers[id] = customer
	return customer, nil
}

// Delete deletes a customer by ID
func (r *MemoryCustomerRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.existsByIDUnsafe(id) {
		return errors.New("customer not found")
	}

	delete(r.customers, id)
	return nil
}

// ExistsByID checks if a customer exists by ID
func (r *MemoryCustomerRepository) ExistsByID(id string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.existsByIDUnsafe(id)
}

// GetByEmail retrieves a customer by email
func (r *MemoryCustomerRepository) GetByEmail(email string) (*model.Customer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	customer := r.getByEmailUnsafe(email)
	if customer == nil {
		return nil, errors.New("customer not found")
	}

	return customer, nil
}

// existsByIDUnsafe checks if a customer exists by ID (without locking)
func (r *MemoryCustomerRepository) existsByIDUnsafe(id string) bool {
	_, exists := r.customers[id]
	return exists
}

// existsByEmailUnsafe checks if a customer exists by email (without locking)
func (r *MemoryCustomerRepository) existsByEmailUnsafe(email string) bool {
	return r.getByEmailUnsafe(email) != nil
}

// getByEmailUnsafe retrieves a customer by email (without locking)
func (r *MemoryCustomerRepository) getByEmailUnsafe(email string) *model.Customer {
	for _, customer := range r.customers {
		if customer.Email == email {
			return customer
		}
	}
	return nil
}

// initSampleData initializes the repository with sample data
func (r *MemoryCustomerRepository) initSampleData() {
	sampleCustomers := []*model.Customer{
		{
			ID:     "customer-456",
			Name:   "John Doe",
			Email:  "john.doe@example.com",
			Phone:  "+1-555-0123",
			Active: true,
			Status: model.StatusActive,
		},
		{
			ID:     "customer-001",
			Name:   "Jane Smith",
			Email:  "jane.smith@example.com",
			Phone:  "+1-555-0124",
			Active: true,
			Status: model.StatusActive,
		},
		{
			ID:     "customer-002",
			Name:   "Bob Johnson",
			Email:  "bob.johnson@example.com",
			Phone:  "+1-555-0125",
			Active: true,
			Status: model.StatusActive,
		},
		{
			ID:     "customer-003",
			Name:   "Alice Brown",
			Email:  "alice.brown@example.com",
			Phone:  "+1-555-0126",
			Active: true,
			Status: model.StatusActive,
		},
		{
			ID:     "customer-004",
			Name:   "Charlie Wilson",
			Email:  "charlie.wilson@example.com",
			Phone:  "+1-555-0127",
			Active: true,
			Status: model.StatusActive,
		},
		{
			ID:     "customer-inactive",
			Name:   "Inactive User",
			Email:  "inactive@example.com",
			Phone:  "+1-555-0128",
			Active: false,
			Status: model.StatusInactive,
		},
		{
			ID:     "customer-blocked",
			Name:   "Blocked User",
			Email:  "blocked@example.com",
			Phone:  "+1-555-0129",
			Active: false,
			Status: model.StatusBlocked,
		},
		{
			ID:     "customer-pending",
			Name:   "Pending User",
			Email:  "pending@example.com",
			Phone:  "+1-555-0130",
			Active: false,
			Status: model.StatusPending,
		},
	}

	for _, customer := range sampleCustomers {
		r.customers[customer.ID] = customer
	}
}
