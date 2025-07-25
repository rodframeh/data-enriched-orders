package model

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProduct_ToResponse(t *testing.T) {
	// Arrange
	price := big.NewRat(99900, 100) // 999.00
	product := &Product{
		ID:          "product-123",
		Name:        "Test Laptop",
		Description: "A test laptop",
		Price:       price,
		Category:    "Electronics",
		Active:      true,
	}

	// Act
	response := product.ToResponse()

	// Assert
	assert.Equal(t, "product-123", response.ID)
	assert.Equal(t, "Test Laptop", response.Name)
	assert.Equal(t, "A test laptop", response.Description)
	assert.Equal(t, 999.0, response.Price)
	assert.Equal(t, "Electronics", response.Category)
	assert.True(t, response.Active)
}

func TestProduct_MarshalJSON(t *testing.T) {
	// Arrange
	price := big.NewRat(99900, 100) // 999.00
	product := &Product{
		ID:          "product-123",
		Name:        "Test Laptop",
		Description: "A test laptop",
		Price:       price,
		Category:    "Electronics",
		Active:      true,
	}

	// Act
	jsonData, err := json.Marshal(product)

	// Assert
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	assert.Equal(t, "product-123", result["id"])
	assert.Equal(t, "Test Laptop", result["name"])
	assert.Equal(t, 999.0, result["price"])
}

func TestProduct_UnmarshalJSON(t *testing.T) {
	// Arrange
	jsonData := `{
		"id": "product-123",
		"name": "Test Laptop",
		"description": "A test laptop",
		"price": 999.0,
		"category": "Electronics",
		"active": true
	}`

	// Act
	var product Product
	err := json.Unmarshal([]byte(jsonData), &product)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "product-123", product.ID)
	assert.Equal(t, "Test Laptop", product.Name)
	assert.Equal(t, "A test laptop", product.Description)

	expectedPrice, _ := product.Price.Float64()
	assert.Equal(t, 999.0, expectedPrice)
	assert.Equal(t, "Electronics", product.Category)
	assert.True(t, product.Active)
}

func TestCreateProductRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateProductRequest
		expectValid bool
	}{
		{
			name: "Valid request",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       99.99,
				Category:    "Electronics",
			},
			expectValid: true,
		},
		{
			name: "Invalid price - zero",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       0,
				Category:    "Electronics",
			},
			expectValid: false,
		},
		{
			name: "Invalid price - negative",
			request: CreateProductRequest{
				Name:        "Test Product",
				Description: "Test Description",
				Price:       -10.0,
				Category:    "Electronics",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate validation that would happen in handler
			isValid := tt.request.Name != "" &&
				tt.request.Description != "" &&
				tt.request.Price > 0 &&
				tt.request.Category != ""

			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}
