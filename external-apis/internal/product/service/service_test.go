package service

import (
	"errors"
	"math/big"
	"testing"

	"external-apis/internal/product/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetByID(id string) (*model.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepository) GetAll() ([]*model.Product, error) {
	args := m.Called()
	return args.Get(0).([]*model.Product), args.Error(1)
}

func (m *MockProductRepository) Create(product *model.Product) (*model.Product, error) {
	args := m.Called(product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepository) Update(id string, product *model.Product) (*model.Product, error) {
	args := m.Called(id, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) ExistsByID(id string) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func TestProductService_GetProductByID(t *testing.T) {
	t.Run("Get existing product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		expectedProduct := &model.Product{
			ID:          "product-123",
			Name:        "Test Product",
			Description: "Test Description",
			Price:       big.NewRat(9999, 100),
			Category:    "Electronics",
			Active:      true,
		}

		mockRepo.On("GetByID", "product-123").Return(expectedProduct, nil)

		// Act
		result, err := service.GetProductByID("product-123")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "product-123", result.ID)
		assert.Equal(t, "Test Product", result.Name)
		assert.Equal(t, 99.99, result.Price)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Get non-existing product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		mockRepo.On("GetByID", "non-existing").Return(nil, errors.New("product not found"))

		// Act
		result, err := service.GetProductByID("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "product not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_GetAllProducts(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo)

	expectedProducts := []*model.Product{
		{
			ID:          "product-1",
			Name:        "Product 1",
			Description: "Description 1",
			Price:       big.NewRat(1000, 100),
			Category:    "Electronics",
			Active:      true,
		},
		{
			ID:          "product-2",
			Name:        "Product 2",
			Description: "Description 2",
			Price:       big.NewRat(2000, 100),
			Category:    "Electronics",
			Active:      true,
		},
	}

	mockRepo.On("GetAll").Return(expectedProducts, nil)

	// Act
	result, err := service.GetAllProducts()

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "product-1", result[0].ID)
	assert.Equal(t, "product-2", result[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestProductService_CreateProduct(t *testing.T) {
	t.Run("Create valid product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		request := model.CreateProductRequest{
			Name:        "New Product",
			Description: "New Description",
			Price:       99.99,
			Category:    "Electronics",
		}

		expectedProduct := &model.Product{
			ID:          "generated-id",
			Name:        "New Product",
			Description: "New Description",
			Price:       big.NewRat(9999, 100),
			Category:    "Electronics",
			Active:      true,
		}

		mockRepo.On("Create", mock.MatchedBy(func(p *model.Product) bool {
			return p.Name == "New Product" && p.Active == true
		})).Return(expectedProduct, nil)

		// Act
		result, err := service.CreateProduct(request)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "generated-id", result.ID)
		assert.Equal(t, "New Product", result.Name)
		assert.True(t, result.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create product with invalid price", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		request := model.CreateProductRequest{
			Name:        "Invalid Product",
			Description: "Invalid Description",
			Price:       -10.0, // Invalid price
			Category:    "Electronics",
		}

		// Act
		result, err := service.CreateProduct(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "price must be greater than 0", err.Error())
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Create product with zero price", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		request := model.CreateProductRequest{
			Name:        "Zero Price Product",
			Description: "Zero Price Description",
			Price:       0.0, // Invalid price
			Category:    "Electronics",
		}

		// Act
		result, err := service.CreateProduct(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "price must be greater than 0", err.Error())
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	t.Run("Update existing product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		existingProduct := &model.Product{
			ID:          "product-123",
			Name:        "Old Name",
			Description: "Old Description",
			Price:       big.NewRat(5000, 100),
			Category:    "Electronics",
			Active:      true,
		}

		newName := "New Name"
		newPrice := 199.99
		updateRequest := model.UpdateProductRequest{
			Name:  &newName,
			Price: &newPrice,
		}

		updatedProduct := &model.Product{
			ID:          "product-123",
			Name:        "New Name",
			Description: "Old Description",
			Price:       big.NewRat(19999, 100),
			Category:    "Electronics",
			Active:      true,
		}

		mockRepo.On("GetByID", "product-123").Return(existingProduct, nil)
		mockRepo.On("Update", "product-123", mock.MatchedBy(func(p *model.Product) bool {
			return p.Name == "New Name"
		})).Return(updatedProduct, nil)

		// Act
		result, err := service.UpdateProduct("product-123", updateRequest)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "New Name", result.Name)
		assert.Equal(t, 199.99, result.Price)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update with invalid price", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		existingProduct := &model.Product{
			ID:          "product-123",
			Name:        "Test Product",
			Description: "Test Description",
			Price:       big.NewRat(5000, 100),
			Category:    "Electronics",
			Active:      true,
		}

		invalidPrice := -50.0
		updateRequest := model.UpdateProductRequest{
			Price: &invalidPrice,
		}

		mockRepo.On("GetByID", "product-123").Return(existingProduct, nil)

		// Act
		result, err := service.UpdateProduct("product-123", updateRequest)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "price must be greater than 0", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update non-existing product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		newName := "New Name"
		updateRequest := model.UpdateProductRequest{
			Name: &newName,
		}

		mockRepo.On("GetByID", "non-existing").Return(nil, errors.New("product not found"))

		// Act
		result, err := service.UpdateProduct("non-existing", updateRequest)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "product not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	t.Run("Delete existing product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		mockRepo.On("Delete", "product-123").Return(nil)

		// Act
		err := service.DeleteProduct("product-123")

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete non-existing product", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockProductRepository)
		service := NewProductService(mockRepo)

		mockRepo.On("Delete", "non-existing").Return(errors.New("product not found"))

		// Act
		err := service.DeleteProduct("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "product not found", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestProductService_ProductExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	service := NewProductService(mockRepo)

	t.Run("Product exists", func(t *testing.T) {
		mockRepo.On("ExistsByID", "existing-product").Return(true)

		// Act
		exists := service.ProductExists("existing-product")

		// Assert
		assert.True(t, exists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Product does not exist", func(t *testing.T) {
		mockRepo.On("ExistsByID", "non-existing-product").Return(false)

		// Act
		exists := service.ProductExists("non-existing-product")

		// Assert
		assert.False(t, exists)
		mockRepo.AssertExpectations(t)
	})
}
