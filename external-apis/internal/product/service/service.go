package service

import (
	"errors"
	"math/big"

	"external-apis/internal/product/model"
	"external-apis/internal/product/repository"
	"github.com/sirupsen/logrus"
)

// ProductService defines the interface for product business logic
type ProductService interface {
	GetProductByID(id string) (*model.ProductResponse, error)
	GetAllProducts() ([]*model.ProductResponse, error)
	CreateProduct(req model.CreateProductRequest) (*model.ProductResponse, error)
	UpdateProduct(id string, req model.UpdateProductRequest) (*model.ProductResponse, error)
	DeleteProduct(id string) error
	ProductExists(id string) bool
}

// productService implements ProductService
type productService struct {
	repo repository.ProductRepository
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{
		repo: repo,
	}
}

// GetProductByID retrieves a product by ID
func (s *productService) GetProductByID(id string) (*model.ProductResponse, error) {
	logrus.WithField("product_id", id).Debug("Getting product by ID")

	product, err := s.repo.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to get product")
		return nil, err
	}

	response := product.ToResponse()
	logrus.WithField("product_id", id).Debug("Successfully retrieved product")

	return &response, nil
}

// GetAllProducts retrieves all products
func (s *productService) GetAllProducts() ([]*model.ProductResponse, error) {
	logrus.Debug("Getting all products")

	products, err := s.repo.GetAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get all products")
		return nil, err
	}

	responses := make([]*model.ProductResponse, len(products))
	for i, product := range products {
		response := product.ToResponse()
		responses[i] = &response
	}

	logrus.WithField("count", len(responses)).Debug("Successfully retrieved all products")
	return responses, nil
}

// CreateProduct creates a new product
func (s *productService) CreateProduct(req model.CreateProductRequest) (*model.ProductResponse, error) {
	logrus.WithFields(logrus.Fields{
		"name":     req.Name,
		"category": req.Category,
		"price":    req.Price,
	}).Debug("Creating new product")

	// Validate price
	if req.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	// Create product model
	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       big.NewRat(1, 1),
		Category:    req.Category,
		Active:      true, // New products are active by default
	}

	// Set price as rational number
	product.Price.SetFloat64(req.Price)

	// Save product
	createdProduct, err := s.repo.Create(product)
	if err != nil {
		logrus.WithError(err).Error("Failed to create product")
		return nil, err
	}

	response := createdProduct.ToResponse()
	logrus.WithField("product_id", createdProduct.ID).Info("Successfully created product")

	return &response, nil
}

// UpdateProduct updates an existing product
func (s *productService) UpdateProduct(id string, req model.UpdateProductRequest) (*model.ProductResponse, error) {
	logrus.WithField("product_id", id).Debug("Updating product")

	// Get existing product
	existingProduct, err := s.repo.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Product not found for update")
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		existingProduct.Name = *req.Name
	}
	if req.Description != nil {
		existingProduct.Description = *req.Description
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, errors.New("price must be greater than 0")
		}
		existingProduct.Price.SetFloat64(*req.Price)
	}
	if req.Category != nil {
		existingProduct.Category = *req.Category
	}
	if req.Active != nil {
		existingProduct.Active = *req.Active
	}

	// Save updated product
	updatedProduct, err := s.repo.Update(id, existingProduct)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to update product")
		return nil, err
	}

	response := updatedProduct.ToResponse()
	logrus.WithField("product_id", id).Info("Successfully updated product")

	return &response, nil
}

// DeleteProduct deletes a product
func (s *productService) DeleteProduct(id string) error {
	logrus.WithField("product_id", id).Debug("Deleting product")

	err := s.repo.Delete(id)
	if err != nil {
		logrus.WithError(err).WithField("product_id", id).Error("Failed to delete product")
		return err
	}

	logrus.WithField("product_id", id).Info("Successfully deleted product")
	return nil
}

// ProductExists checks if a product exists
func (s *productService) ProductExists(id string) bool {
	return s.repo.ExistsByID(id)
}
