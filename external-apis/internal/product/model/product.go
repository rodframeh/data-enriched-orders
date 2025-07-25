package model

import (
	"encoding/json"
	"math/big"
)

// Product represents a product in the catalog
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       *big.Rat `json:"price"`
	Category    string   `json:"category"`
	Active      bool     `json:"active"`
}

// ProductResponse represents the API response for a product
type ProductResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Active      bool    `json:"active"`
}

// ToResponse converts a Product to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	priceFloat, _ := p.Price.Float64()
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       priceFloat,
		Category:    p.Category,
		Active:      p.Active,
	}
}

// MarshalJSON custom marshaling for Product
func (p *Product) MarshalJSON() ([]byte, error) {
	type Alias Product
	priceFloat, _ := p.Price.Float64()

	return json.Marshal(&struct {
		*Alias
		Price float64 `json:"price"`
	}{
		Alias: (*Alias)(p),
		Price: priceFloat,
	})
}

// UnmarshalJSON custom unmarshaling for Product
func (p *Product) UnmarshalJSON(data []byte) error {
	type Alias Product
	aux := &struct {
		*Alias
		Price float64 `json:"price"`
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	p.Price = big.NewRat(1, 1)
	p.Price.SetFloat64(aux.Price)

	return nil
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Category    *string  `json:"category,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}
