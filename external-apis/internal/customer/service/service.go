package service

import (
	"errors"
	"regexp"

	"external-apis/internal/customer/model"
	"external-apis/internal/customer/repository"
	"github.com/sirupsen/logrus"
)

// CustomerService defines the interface for customer business logic
type CustomerService interface {
	GetCustomerByID(id string) (*model.CustomerResponse, error)
	GetAllCustomers() ([]*model.CustomerResponse, error)
	CreateCustomer(req model.CreateCustomerRequest) (*model.CustomerResponse, error)
	UpdateCustomer(id string, req model.UpdateCustomerRequest) (*model.CustomerResponse, error)
	DeleteCustomer(id string) error
	CustomerExists(id string) bool
	GetCustomerByEmail(email string) (*model.CustomerResponse, error)
}

// customerService implements CustomerService
type customerService struct {
	repo repository.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(repo repository.CustomerRepository) CustomerService {
	return &customerService{
		repo: repo,
	}
}

// GetCustomerByID retrieves a customer by ID
func (s *customerService) GetCustomerByID(id string) (*model.CustomerResponse, error) {
	logrus.WithField("customer_id", id).Debug("Getting customer by ID")

	customer, err := s.repo.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", id).Error("Failed to get customer")
		return nil, err
	}

	response := customer.ToResponse()
	logrus.WithField("customer_id", id).Debug("Successfully retrieved customer")

	return &response, nil
}

// GetAllCustomers retrieves all customers
func (s *customerService) GetAllCustomers() ([]*model.CustomerResponse, error) {
	logrus.Debug("Getting all customers")

	customers, err := s.repo.GetAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get all customers")
		return nil, err
	}

	responses := make([]*model.CustomerResponse, len(customers))
	for i, customer := range customers {
		response := customer.ToResponse()
		responses[i] = &response
	}

	logrus.WithField("count", len(responses)).Debug("Successfully retrieved all customers")
	return responses, nil
}

// CreateCustomer creates a new customer
func (s *customerService) CreateCustomer(req model.CreateCustomerRequest) (*model.CustomerResponse, error) {
	logrus.WithFields(logrus.Fields{
		"name":  req.Name,
		"email": req.Email,
		"phone": req.Phone,
	}).Debug("Creating new customer")

	// Validate email format
	if !isValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Validate phone format
	if !isValidPhone(req.Phone) {
		return nil, errors.New("invalid phone format")
	}

	// Create customer model
	customer := &model.Customer{
		Name:   req.Name,
		Email:  req.Email,
		Phone:  req.Phone,
		Active: true,               // New customers are active by default
		Status: model.StatusActive, // New customers start with active status
	}

	// Save customer
	createdCustomer, err := s.repo.Create(customer)
	if err != nil {
		logrus.WithError(err).Error("Failed to create customer")
		return nil, err
	}

	response := createdCustomer.ToResponse()
	logrus.WithField("customer_id", createdCustomer.ID).Info("Successfully created customer")

	return &response, nil
}

// UpdateCustomer updates an existing customer
func (s *customerService) UpdateCustomer(id string, req model.UpdateCustomerRequest) (*model.CustomerResponse, error) {
	logrus.WithField("customer_id", id).Debug("Updating customer")

	// Get existing customer
	existingCustomer, err := s.repo.GetByID(id)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", id).Error("Customer not found for update")
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		existingCustomer.Name = *req.Name
	}
	if req.Email != nil {
		if !isValidEmail(*req.Email) {
			return nil, errors.New("invalid email format")
		}
		existingCustomer.Email = *req.Email
	}
	if req.Phone != nil {
		if !isValidPhone(*req.Phone) {
			return nil, errors.New("invalid phone format")
		}
		existingCustomer.Phone = *req.Phone
	}
	if req.Active != nil {
		existingCustomer.Active = *req.Active
	}
	if req.Status != nil {
		if !req.Status.IsValid() {
			return nil, errors.New("invalid customer status")
		}
		existingCustomer.Status = *req.Status
	}

	// Save updated customer
	updatedCustomer, err := s.repo.Update(id, existingCustomer)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", id).Error("Failed to update customer")
		return nil, err
	}

	response := updatedCustomer.ToResponse()
	logrus.WithField("customer_id", id).Info("Successfully updated customer")

	return &response, nil
}

// DeleteCustomer deletes a customer
func (s *customerService) DeleteCustomer(id string) error {
	logrus.WithField("customer_id", id).Debug("Deleting customer")

	err := s.repo.Delete(id)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", id).Error("Failed to delete customer")
		return err
	}

	logrus.WithField("customer_id", id).Info("Successfully deleted customer")
	return nil
}

// CustomerExists checks if a customer exists
func (s *customerService) CustomerExists(id string) bool {
	return s.repo.ExistsByID(id)
}

// GetCustomerByEmail retrieves a customer by email
func (s *customerService) GetCustomerByEmail(email string) (*model.CustomerResponse, error) {
	logrus.WithField("email", email).Debug("Getting customer by email")

	customer, err := s.repo.GetByEmail(email)
	if err != nil {
		logrus.WithError(err).WithField("email", email).Error("Failed to get customer by email")
		return nil, err
	}

	response := customer.ToResponse()
	logrus.WithField("email", email).Debug("Successfully retrieved customer by email")

	return &response, nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidPhone validates phone format (basic validation)
func isValidPhone(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}
