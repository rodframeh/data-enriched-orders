package service

import (
	"errors"
	"testing"

	"external-apis/internal/customer/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCustomerRepository is a mock implementation of CustomerRepository
type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) GetByID(id string) (*model.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetAll() ([]*model.Customer, error) {
	args := m.Called()
	return args.Get(0).([]*model.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Create(customer *model.Customer) (*model.Customer, error) {
	args := m.Called(customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Update(id string, customer *model.Customer) (*model.Customer, error) {
	args := m.Called(id, customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Customer), args.Error(1)
}

func (m *MockCustomerRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCustomerRepository) ExistsByID(id string) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func (m *MockCustomerRepository) GetByEmail(email string) (*model.Customer, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Customer), args.Error(1)
}

func TestCustomerService_GetCustomerByID(t *testing.T) {
	t.Run("Get existing customer", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		expectedCustomer := &model.Customer{
			ID:     "customer-123",
			Name:   "John Doe",
			Email:  "john.doe@example.com",
			Phone:  "+15550123",
			Active: true,
			Status: model.StatusActive,
		}

		mockRepo.On("GetByID", "customer-123").Return(expectedCustomer, nil)

		// Act
		result, err := service.GetCustomerByID("customer-123")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "customer-123", result.ID)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, "john.doe@example.com", result.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Get non-existing customer", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		mockRepo.On("GetByID", "non-existing").Return(nil, errors.New("customer not found"))

		// Act
		result, err := service.GetCustomerByID("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "customer not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_GetCustomerByEmail(t *testing.T) {
	t.Run("Get customer by existing email", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		expectedCustomer := &model.Customer{
			ID:     "customer-123",
			Name:   "John Doe",
			Email:  "john.doe@example.com",
			Phone:  "+15550123",
			Active: true,
			Status: model.StatusActive,
		}

		mockRepo.On("GetByEmail", "john.doe@example.com").Return(expectedCustomer, nil)

		// Act
		result, err := service.GetCustomerByEmail("john.doe@example.com")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "customer-123", result.ID)
		assert.Equal(t, "john.doe@example.com", result.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Get customer by non-existing email", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		mockRepo.On("GetByEmail", "nonexisting@example.com").Return(nil, errors.New("customer not found"))

		// Act
		result, err := service.GetCustomerByEmail("nonexisting@example.com")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_CreateCustomer(t *testing.T) {
	t.Run("Create valid customer", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		request := model.CreateCustomerRequest{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			Phone: "+15550123", // Sin guiones para coincidir con el regex
		}

		expectedCustomer := &model.Customer{
			ID:     "generated-id",
			Name:   "John Doe",
			Email:  "john.doe@example.com",
			Phone:  "+15550123",
			Active: true,
			Status: model.StatusActive,
		}

		mockRepo.On("Create", mock.MatchedBy(func(c *model.Customer) bool {
			return c.Name == "John Doe" &&
				c.Email == "john.doe@example.com" &&
				c.Phone == "+15550123" &&
				c.Active == true &&
				c.Status == model.StatusActive
		})).Return(expectedCustomer, nil)

		// Act
		result, err := service.CreateCustomer(request)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "generated-id", result.ID)
		assert.Equal(t, "John Doe", result.Name)
		assert.True(t, result.Active)
		assert.Equal(t, model.StatusActive, result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create customer with invalid email", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		request := model.CreateCustomerRequest{
			Name:  "John Doe",
			Email: "invalid-email", // Invalid email format
			Phone: "+15550123",
		}

		// Act
		result, err := service.CreateCustomer(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid email format", err.Error())
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Create customer with invalid phone", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		request := model.CreateCustomerRequest{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			Phone: "invalid-phone", // Invalid phone format
		}

		// Act
		result, err := service.CreateCustomer(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid phone format", err.Error())
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestCustomerService_UpdateCustomer(t *testing.T) {
	t.Run("Update existing customer", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		existingCustomer := &model.Customer{
			ID:     "customer-123",
			Name:   "Old Name",
			Email:  "old@example.com",
			Phone:  "+15550000",
			Active: true,
			Status: model.StatusActive,
		}

		newName := "New Name"
		newStatus := model.StatusInactive
		updateRequest := model.UpdateCustomerRequest{
			Name:   &newName,
			Status: &newStatus,
		}

		updatedCustomer := &model.Customer{
			ID:     "customer-123",
			Name:   "New Name",
			Email:  "old@example.com",
			Phone:  "+15550000",
			Active: true,
			Status: model.StatusInactive,
		}

		mockRepo.On("GetByID", "customer-123").Return(existingCustomer, nil)
		mockRepo.On("Update", "customer-123", mock.MatchedBy(func(c *model.Customer) bool {
			return c.Name == "New Name" && c.Status == model.StatusInactive
		})).Return(updatedCustomer, nil)

		// Act
		result, err := service.UpdateCustomer("customer-123", updateRequest)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "New Name", result.Name)
		assert.Equal(t, model.StatusInactive, result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update with invalid email", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		existingCustomer := &model.Customer{
			ID:     "customer-123",
			Name:   "John Doe",
			Email:  "john@example.com",
			Phone:  "+15550123",
			Active: true,
			Status: model.StatusActive,
		}

		invalidEmail := "invalid-email"
		updateRequest := model.UpdateCustomerRequest{
			Email: &invalidEmail,
		}

		mockRepo.On("GetByID", "customer-123").Return(existingCustomer, nil)

		// Act
		result, err := service.UpdateCustomer("customer-123", updateRequest)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid email format", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update with invalid status", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		existingCustomer := &model.Customer{
			ID:     "customer-123",
			Name:   "John Doe",
			Email:  "john@example.com",
			Phone:  "+15550123",
			Active: true,
			Status: model.StatusActive,
		}

		invalidStatus := model.CustomerStatus("INVALID")
		updateRequest := model.UpdateCustomerRequest{
			Status: &invalidStatus,
		}

		mockRepo.On("GetByID", "customer-123").Return(existingCustomer, nil)

		// Act
		result, err := service.UpdateCustomer("customer-123", updateRequest)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid customer status", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_DeleteCustomer(t *testing.T) {
	t.Run("Delete existing customer", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		mockRepo.On("Delete", "customer-123").Return(nil)

		// Act
		err := service.DeleteCustomer("customer-123")

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete non-existing customer", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockCustomerRepository)
		service := NewCustomerService(mockRepo)

		mockRepo.On("Delete", "non-existing").Return(errors.New("customer not found"))

		// Act
		err := service.DeleteCustomer("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "customer not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_CustomerExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockCustomerRepository)
	service := NewCustomerService(mockRepo)

	t.Run("Customer exists", func(t *testing.T) {
		mockRepo.On("ExistsByID", "existing-customer").Return(true)

		// Act
		exists := service.CustomerExists("existing-customer")

		// Assert
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Customer does not exist", func(t *testing.T) {
		mockRepo.On("ExistsByID", "non-existing-customer").Return(false)

		// Act
		exists := service.CustomerExists("non-existing-customer")

		// Assert
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_GetAllCustomers(t *testing.T) {
	// Arrange
	mockRepo := new(MockCustomerRepository)
	service := NewCustomerService(mockRepo)

	expectedCustomers := []*model.Customer{
		{
			ID:     "customer-1",
			Name:   "Customer 1",
			Email:  "customer1@example.com",
			Phone:  "+15550001",
			Active: true,
			Status: model.StatusActive,
		},
		{
			ID:     "customer-2",
			Name:   "Customer 2",
			Email:  "customer2@example.com",
			Phone:  "+15550002",
			Active: false,
			Status: model.StatusInactive,
		},
	}

	mockRepo.On("GetAll").Return(expectedCustomers, nil)

	// Act
	result, err := service.GetAllCustomers()

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "customer-1", result[0].ID)
	assert.Equal(t, "customer-2", result[1].ID)
	mockRepo.AssertExpectations(t)
}

// Test email validation function
func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid email", "test@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Valid email with numbers", "user123@example.com", true},
		{"Valid email with special chars", "user.name+tag@example.com", true},
		{"Invalid email - no @", "userexample.com", false},
		{"Invalid email - no domain", "user@", false},
		{"Invalid email - no TLD", "user@example", false},
		{"Invalid email - spaces", "user @example.com", false},
		{"Invalid email - empty", "", false},
		{"Invalid email - only @", "@", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test phone validation function
func TestPhoneValidation(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"Valid US phone with +", "+15550123", true},
		{"Valid international phone", "+442079460958", true},
		{"Valid phone with more digits", "+123456789012345", true},
		{"Valid phone without + (regex allows it)", "15550123", true}, // El regex \+? hace el + opcional
		{"Valid short phone without +", "1555", true},                 // El regex permite \d{1,14}
		{"Valid short phone with +", "+1555", true},                   // 3 dígitos adicionales es válido
		{"Invalid phone - starts with 0", "+05550123", false},
		{"Invalid phone - starts with 0 without +", "05550123", false},
		{"Invalid phone - letters", "+1ABCDEFG", false},
		{"Invalid phone - letters without +", "1ABCDEFG", false},
		{"Invalid phone - empty", "", false},
		{"Invalid phone - only +", "+", false},
		{"Invalid phone - with dashes", "+1-555-0123", false},           // El regex no acepta guiones
		{"Invalid phone - with spaces", "+1 555 0123", false},           // No acepta espacios
		{"Invalid phone - with parentheses", "+1(555)0123", false},      // No acepta paréntesis
		{"Invalid phone - too many digits", "+1234567890123456", false}, // Más de 15 dígitos total (1 + 14)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPhone(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}
