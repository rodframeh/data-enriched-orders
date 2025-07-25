package repository

import (
	"math/big"
	"testing"

	"external-apis/internal/product/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryProductRepository_GetByID(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()

	t.Run("Get existing product", func(t *testing.T) {
		// Act
		product, err := repo.GetByID("product-789")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "product-789", product.ID)
		assert.Equal(t, "Laptop", product.Name)
		assert.True(t, product.Active)
	})

	t.Run("Get non-existing product", func(t *testing.T) {
		// Act
		product, err := repo.GetByID("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Equal(t, "product not found", err.Error())
	})
}

func TestMemoryProductRepository_GetAll(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()

	// Act
	products, err := repo.GetAll()

	// Assert
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 10) // At least the sample data

	// Verify we have the expected sample products
	foundLaptop := false
	for _, product := range products {
		if product.ID == "product-789" && product.Name == "Laptop" {
			foundLaptop = true
			break
		}
	}
	assert.True(t, foundLaptop, "Should contain the sample laptop product")
}

func TestMemoryProductRepository_Create(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()
	newProduct := &model.Product{
		Name:        "New Product",
		Description: "A new test product",
		Price:       big.NewRat(4999, 100), // 49.99
		Category:    "Test",
		Active:      true,
	}

	t.Run("Create new product", func(t *testing.T) {
		// Act
		created, err := repo.Create(newProduct)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "New Product", created.Name)

		// Verify it was actually stored
		retrieved, err := repo.GetByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
	})

	t.Run("Create product with existing ID", func(t *testing.T) {
		// Arrange
		existingProduct := &model.Product{
			ID:          "product-789", // This ID already exists
			Name:        "Duplicate",
			Description: "Duplicate product",
			Price:       big.NewRat(1000, 100),
			Category:    "Test",
			Active:      true,
		}

		// Act
		created, err := repo.Create(existingProduct)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Equal(t, "product already exists", err.Error())
	})
}

func TestMemoryProductRepository_Update(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()

	t.Run("Update existing product", func(t *testing.T) {
		// Arrange
		updatedProduct := &model.Product{
			ID:          "product-789",
			Name:        "Updated Laptop",
			Description: "Updated description",
			Price:       big.NewRat(119900, 100), // 1199.00
			Category:    "Electronics",
			Active:      true,
		}

		// Act
		result, err := repo.Update("product-789", updatedProduct)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "Updated Laptop", result.Name)
		assert.Equal(t, "Updated description", result.Description)

		// Verify the update was persisted
		retrieved, err := repo.GetByID("product-789")
		require.NoError(t, err)
		assert.Equal(t, "Updated Laptop", retrieved.Name)
	})

	t.Run("Update non-existing product", func(t *testing.T) {
		// Arrange
		product := &model.Product{
			Name:        "Non-existing",
			Description: "Does not exist",
			Price:       big.NewRat(1000, 100),
			Category:    "Test",
			Active:      true,
		}

		// Act
		result, err := repo.Update("non-existing", product)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "product not found", err.Error())
	})
}

func TestMemoryProductRepository_Delete(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()

	t.Run("Delete existing product", func(t *testing.T) {
		// Act
		err := repo.Delete("product-001")

		// Assert
		require.NoError(t, err)

		// Verify it was deleted
		product, err := repo.GetByID("product-001")
		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("Delete non-existing product", func(t *testing.T) {
		// Act
		err := repo.Delete("non-existing")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "product not found", err.Error())
	})
}

func TestMemoryProductRepository_ExistsByID(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()

	t.Run("Existing product", func(t *testing.T) {
		// Act
		exists := repo.ExistsByID("product-789")

		// Assert
		assert.True(t, exists)
	})

	t.Run("Non-existing product", func(t *testing.T) {
		// Act
		exists := repo.ExistsByID("non-existing")

		// Assert
		assert.False(t, exists)
	})
}

func TestMemoryProductRepository_ConcurrentAccess(t *testing.T) {
	// Arrange
	repo := NewMemoryProductRepository()

	t.Run("Concurrent reads", func(t *testing.T) {
		// Act & Assert - should not panic
		for i := 0; i < 10; i++ {
			go func() {
				_, _ = repo.GetByID("product-789")
			}()
		}
	})

	t.Run("Concurrent writes", func(t *testing.T) {
		// Act & Assert - should not panic
		for i := 0; i < 10; i++ {
			go func(index int) {
				product := &model.Product{
					Name:        "Concurrent Product",
					Description: "Test concurrent access",
					Price:       big.NewRat(1000, 100),
					Category:    "Test",
					Active:      true,
				}
				_, _ = repo.Create(product)
			}(i)
		}
	})
}
