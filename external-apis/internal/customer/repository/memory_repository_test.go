package repository

import (
	"testing"

	"external-apis/internal/customer/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCustomerRepository_GetByID(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Get existing customer", func(t *testing.T) {
		// Act
		customer, err := repo.GetByID("customer-456")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "customer-456", customer.ID)
		assert.Equal(t, "John Doe", customer.Name)
		assert.Equal(t, "john.doe@example.com", customer.Email)
		assert.True(t, customer.Active)
		assert.Equal(t, model.StatusActive, customer.Status)
	})

	t.Run("Get non-existing customer", func(t *testing.T) {
		// Act
		customer, err := repo.GetByID("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.Equal(t, "customer not found", err.Error())
	})
}

func TestMemoryCustomerRepository_GetByEmail(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Get customer by existing email", func(t *testing.T) {
		// Act
		customer, err := repo.GetByEmail("john.doe@example.com")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "customer-456", customer.ID)
		assert.Equal(t, "John Doe", customer.Name)
		assert.Equal(t, "john.doe@example.com", customer.Email)
	})

	t.Run("Get customer by non-existing email", func(t *testing.T) {
		// Act
		customer, err := repo.GetByEmail("nonexisting@example.com")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.Equal(t, "customer not found", err.Error())
	})
}

func TestMemoryCustomerRepository_GetAll(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	// Act
	customers, err := repo.GetAll()

	// Assert
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(customers), 8) // At least the sample data

	// Verify we have customers with different statuses
	statuses := make(map[model.CustomerStatus]int)
	for _, customer := range customers {
		statuses[customer.Status]++
	}

	assert.Greater(t, statuses[model.StatusActive], 0)
	assert.Greater(t, statuses[model.StatusInactive], 0)
	assert.Greater(t, statuses[model.StatusBlocked], 0)
	assert.Greater(t, statuses[model.StatusPending], 0)
}

func TestMemoryCustomerRepository_Create(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Create new customer", func(t *testing.T) {
		// Arrange
		newCustomer := &model.Customer{
			Name:   "New Customer",
			Email:  "new@example.com",
			Phone:  "+1-555-0999",
			Active: true,
			Status: model.StatusActive,
		}

		// Act
		created, err := repo.Create(newCustomer)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "New Customer", created.Name)
		assert.Equal(t, "new@example.com", created.Email)

		// Verify it was actually stored
		retrieved, err := repo.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
	})

	t.Run("Create customer with duplicate email", func(t *testing.T) {
		// Arrange
		duplicateCustomer := &model.Customer{
			Name:   "Duplicate Email",
			Email:  "john.doe@example.com", // This email already exists
			Phone:  "+1-555-0888",
			Active: true,
			Status: model.StatusActive,
		}

		// Act
		created, err := repo.Create(duplicateCustomer)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Equal(t, "customer with this email already exists", err.Error())
	})

	t.Run("Create customer with existing ID", func(t *testing.T) {
		// Arrange
		existingCustomer := &model.Customer{
			ID:     "customer-456", // This ID already exists
			Name:   "Duplicate ID",
			Email:  "duplicate@example.com",
			Phone:  "+1-555-0777",
			Active: true,
			Status: model.StatusActive,
		}

		// Act
		created, err := repo.Create(existingCustomer)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Equal(t, "customer already exists", err.Error())
	})
}

func TestMemoryCustomerRepository_Update(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Update existing customer", func(t *testing.T) {
		// Arrange
		updatedCustomer := &model.Customer{
			ID:     "customer-456",
			Name:   "Updated John Doe",
			Email:  "updated.john@example.com",
			Phone:  "+1-555-9999",
			Active: false,
			Status: model.StatusInactive,
		}

		// Act
		result, err := repo.Update("customer-456", updatedCustomer)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Updated John Doe", result.Name)
		assert.Equal(t, "updated.john@example.com", result.Email)
		assert.False(t, result.Active)
		assert.Equal(t, model.StatusInactive, result.Status)

		// Verify the update was persisted
		retrieved, err := repo.GetByID("customer-456")
		require.NoError(t, err)
		assert.Equal(t, "Updated John Doe", retrieved.Name)
	})

	t.Run("Update with duplicate email", func(t *testing.T) {
		// Arrange
		updatedCustomer := &model.Customer{
			ID:     "customer-001",
			Name:   "Jane Smith",
			Email:  "updated.john@example.com", // This email was used in previous test
			Phone:  "+1-555-0124",
			Active: true,
			Status: model.StatusActive,
		}

		// Act
		result, err := repo.Update("customer-001", updatedCustomer)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "customer with this email already exists", err.Error())
	})

	t.Run("Update non-existing customer", func(t *testing.T) {
		// Arrange
		customer := &model.Customer{
			Name:   "Non-existing",
			Email:  "nonexisting@example.com",
			Phone:  "+1-555-0000",
			Active: true,
			Status: model.StatusActive,
		}

		// Act
		result, err := repo.Update("non-existing", customer)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "customer not found", err.Error())
	})
}

func TestMemoryCustomerRepository_Delete(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Delete existing customer", func(t *testing.T) {
		// Act
		err := repo.Delete("customer-001")

		// Assert
		require.NoError(t, err)

		// Verify it was deleted
		customer, err := repo.GetByID("customer-001")
		assert.Error(t, err)
		assert.Nil(t, customer)
	})

	t.Run("Delete non-existing customer", func(t *testing.T) {
		// Act
		err := repo.Delete("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "customer not found", err.Error())
	})
}

func TestMemoryCustomerRepository_ExistsByID(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Existing customer", func(t *testing.T) {
		// Act
		exists := repo.ExistsByID("customer-456")

		// Assert
		assert.True(t, exists)
	})

	t.Run("Non-existing customer", func(t *testing.T) {
		// Act
		exists := repo.ExistsByID("non-existing")

		// Assert
		assert.False(t, exists)
	})
}

func TestMemoryCustomerRepository_ConcurrentAccess(t *testing.T) {
	// Arrange
	repo := NewMemoryCustomerRepository()

	t.Run("Concurrent reads", func(t *testing.T) {
		// Act & Assert - should not panic
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, _ = repo.GetByID("customer-456")
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("Concurrent writes", func(t *testing.T) {
		// Act & Assert - should not panic
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				customer := &model.Customer{
					Name:   "Concurrent Customer",
					Email:  "concurrent" + string(rune(index)) + "@example.com",
					Phone:  "+1-555-0000",
					Active: true,
					Status: model.StatusActive,
				}
				_, _ = repo.Create(customer)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
