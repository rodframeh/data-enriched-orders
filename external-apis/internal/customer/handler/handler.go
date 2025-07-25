package handler

import (
	"external-apis/internal/customer/model"
	"external-apis/internal/customer/service"
	"external-apis/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CustomerHandler handles HTTP requests for customers
type CustomerHandler struct {
	service service.CustomerService
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(service service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		service: service,
	}
}

// RegisterRoutes registers all customer routes
func (h *CustomerHandler) RegisterRoutes(router *gin.RouterGroup) {
	customers := router.Group("/customers")
	{
		customers.GET("", h.GetAllCustomers)
		customers.GET("/:id", h.GetCustomerByID)
		customers.GET("/email/:email", h.GetCustomerByEmail)
		customers.POST("", h.CreateCustomer)
		customers.PUT("/:id", h.UpdateCustomer)
		customers.DELETE("/:id", h.DeleteCustomer)
	}
}

// GetCustomerByID godoc
// @Summary Get customer by ID
// @Description Get a customer by its ID
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} response.SuccessResponse{data=model.CustomerResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/customers/{id} [get]
func (h *CustomerHandler) GetCustomerByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		response.BadRequest(c, "Customer ID is required")
		return
	}

	logrus.WithFields(logrus.Fields{
		"customer_id": id,
		"request_id":  c.GetString("request_id"),
	}).Info("Getting customer by ID")

	customer, err := h.service.GetCustomerByID(id)
	if err != nil {
		if err.Error() == "customer not found" {
			response.NotFound(c, "Customer not found")
			return
		}

		logrus.WithError(err).WithField("customer_id", id).Error("Failed to get customer")
		response.InternalServerError(c, "Failed to retrieve customer")
		return
	}

	response.OK(c, customer)
}

// GetAllCustomers godoc
// @Summary Get all customers
// @Description Get a list of all customers
// @Tags customers
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]model.CustomerResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/customers [get]
func (h *CustomerHandler) GetAllCustomers(c *gin.Context) {
	logrus.WithField("request_id", c.GetString("request_id")).Info("Getting all customers")

	customers, err := h.service.GetAllCustomers()
	if err != nil {
		logrus.WithError(err).Error("Failed to get all customers")
		response.InternalServerError(c, "Failed to retrieve customers")
		return
	}

	response.OK(c, customers)
}

// GetCustomerByEmail godoc
// @Summary Get customer by email
// @Description Get a customer by its email address
// @Tags customers
// @Accept json
// @Produce json
// @Param email path string true "Customer Email"
// @Success 200 {object} response.SuccessResponse{data=model.CustomerResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/customers/email/{email} [get]
func (h *CustomerHandler) GetCustomerByEmail(c *gin.Context) {
	email := c.Param("email")

	if email == "" {
		response.BadRequest(c, "Customer email is required")
		return
	}

	logrus.WithFields(logrus.Fields{
		"email":      email,
		"request_id": c.GetString("request_id"),
	}).Info("Getting customer by email")

	customer, err := h.service.GetCustomerByEmail(email)
	if err != nil {
		if err.Error() == "customer not found" {
			response.NotFound(c, "Customer not found")
			return
		}

		logrus.WithError(err).WithField("email", email).Error("Failed to get customer by email")
		response.InternalServerError(c, "Failed to retrieve customer")
		return
	}

	response.OK(c, customer)
}

// CreateCustomer godoc
// @Summary Create a new customer
// @Description Create a new customer
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body model.CreateCustomerRequest true "Customer data"
// @Success 201 {object} response.SuccessResponse{data=model.CustomerResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req model.CreateCustomerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body for create customer")
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"name":       req.Name,
		"email":      req.Email,
		"request_id": c.GetString("request_id"),
	}).Info("Creating new customer")

	customer, err := h.service.CreateCustomer(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create customer")

		if err.Error() == "customer already exists" || err.Error() == "customer with this email already exists" {
			response.Conflict(c, err.Error())
			return
		}

		if err.Error() == "invalid email format" || err.Error() == "invalid phone format" {
			response.BadRequest(c, err.Error())
			return
		}

		response.InternalServerError(c, "Failed to create customer")
		return
	}

	response.Created(c, customer)
}

// UpdateCustomer godoc
// @Summary Update a customer
// @Description Update an existing customer
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param customer body model.UpdateCustomerRequest true "Customer data"
// @Success 200 {object} response.SuccessResponse{data=model.CustomerResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		response.BadRequest(c, "Customer ID is required")
		return
	}

	var req model.UpdateCustomerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid request body for update customer")
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	logrus.WithFields(logrus.Fields{
		"customer_id": id,
		"request_id":  c.GetString("request_id"),
	}).Info("Updating customer")

	customer, err := h.service.UpdateCustomer(id, req)
	if err != nil {
		if err.Error() == "customer not found" {
			response.NotFound(c, "Customer not found")
			return
		}

		if err.Error() == "customer with this email already exists" {
			response.Conflict(c, err.Error())
			return
		}

		if err.Error() == "invalid email format" || err.Error() == "invalid phone format" || err.Error() == "invalid customer status" {
			response.BadRequest(c, err.Error())
			return
		}

		logrus.WithError(err).WithField("customer_id", id).Error("Failed to update customer")
		response.InternalServerError(c, "Failed to update customer")
		return
	}

	response.OK(c, customer)
}

// DeleteCustomer godoc
// @Summary Delete a customer
// @Description Delete a customer by ID
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		response.BadRequest(c, "Customer ID is required")
		return
	}

	logrus.WithFields(logrus.Fields{
		"customer_id": id,
		"request_id":  c.GetString("request_id"),
	}).Info("Deleting customer")

	err := h.service.DeleteCustomer(id)
	if err != nil {
		if err.Error() == "customer not found" {
			response.NotFound(c, "Customer not found")
			return
		}

		logrus.WithError(err).WithField("customer_id", id).Error("Failed to delete customer")
		response.InternalServerError(c, "Failed to delete customer")
		return
	}

	response.OK(c, gin.H{"message": "Customer deleted successfully"})
}
