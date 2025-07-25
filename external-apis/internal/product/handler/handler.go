package handler

import (
	"external-apis/internal/product/model"
	"external-apis/internal/product/service"
	"external-apis/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	service service.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

// RegisterRoutes registers all product routes
func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")
	{
		products.GET("", h.GetAllProducts)
		products.GET("/:id", h.GetProductByID)
		products.POST("", h.CreateProduct)
		products.PUT("/:id", h.UpdateProduct)
		products.DELETE("/:id", h.DeleteProduct)
	}
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Get a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		response.BadRequest(c, "Product ID is required")
		return
	}

	logrus.WithFields(logrus.Fields{
		"product_id": id,
		"request_id": c.GetString("request_id"),
	}).Info("Getting product by ID")

	product, err := h.service.GetProductByID(id)
	if err != nil {
		if err.Error() == "product not found" {
			response.NotFound(c, "Product not found")
			return
		}

		logrus.WithError(err).WithField("product_id", id).Error("Failed to get product")
		response.InternalServerError(c, "Failed to retrieve product")
		return
	}

	response.OK(c, product)
}

// GetAllProducts godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags products
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.ProductResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	logrus.WithField("request_id", c.GetString("request_id")).Info("Getting all products")

	products, err := h.service.GetAllProducts()
	if err != nil {
		logrus.WithError(err).Error("Failed to get all products")
		response.InternalServerError(c, "Failed to retrieve products")
		return
	}

	response.OK(c, products)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body model.CreateProductRequest true "Product data"
// @Success 201 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req model.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body for create product")
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"name":       req.Name,
		"category":   req.Category,
		"request_id": c.GetString("request_id"),
	}).Info("Creating new product")

	product, err := h.service.CreateProduct(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create product")

		if err.Error() == "product already exists" {
			response.Conflict(c, "Product already exists")
			return
		}

		response.InternalServerError(c, "Failed to create product")
		return
	}

	response.Created(c, product)
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body model.UpdateProductRequest true "Product data"
// @Success 200 {object} response.SuccessResponse{data=model.ProductResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		response.BadRequest(c, "Product ID is required")
		return
	}

	var req model.UpdateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body for update product")
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"product_id": id,
		"request_id": c.GetString("request_id"),
	}).Info("Updating product")

	product, err := h.service.UpdateProduct(id, req)
	if err != nil {
		if err.Error() == "product not found" {
			response.NotFound(c, "Product not found")
			return
		}

		logrus.WithError(err).WithField("product_id", id).Error("Failed to update product")
		response.InternalServerError(c, "Failed to update product")
		return
	}

	response.OK(c, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		response.BadRequest(c, "Product ID is required")
		return
	}

	logrus.WithFields(logrus.Fields{
		"product_id": id,
		"request_id": c.GetString("request_id"),
	}).Info("Deleting product")

	err := h.service.DeleteProduct(id)
	if err != nil {
		if err.Error() == "product not found" {
			response.NotFound(c, "Product not found")
			return
		}

		logrus.WithError(err).WithField("product_id", id).Error("Failed to delete product")
		response.InternalServerError(c, "Failed to delete product")
		return
	}

	response.OK(c, gin.H{"message": "Product deleted successfully"})
}
