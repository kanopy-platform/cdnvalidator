package v1beta1

import (
	"errors"
	"fmt"
	"time"
)

type VanityDistributionName string

type Distribution struct {
	DistributionID string `json:"-"`
	PathPrefix     string `json:"pathPrefix"`
}

type InvalidationMeta struct {
	// The Status of the invalidation request
	Status string `json:"status"`
}

// swagger:model InvalidationResponse
type InvalidationResponse struct {
	InvalidationMeta
	// The ID of the Invalidation Request
	ID string `json:"id,omitempty"`

	// The Created time of invalidation
	Created time.Time `json:"createTime"`

	// The Paths array requested for invalidation
	Paths []string `json:"paths,omitempty"`
}

// swagger:model DistributionResponse
type DistributionsResponse struct {
	// The Distributions a user is entitled to perform invalidations against.
	Distributions map[VanityDistributionName]Distribution `json:"distributions"`
}

// swagger:model InvalidationRequest
type InvalidationRequest struct {
	// The Paths to submit for invalidation
	Paths []string `json:"paths"`
}

// swagger:parameters submit-invalidation
type _ struct {
	// The Name of the distribution
	// in:path
	Name string
	// The body to create the invalidation
	// in:body
	// required: true
	Body InvalidationRequest
}

// swagger:parameters get-invalidation
type _ struct {
	// The Name of the distribution
	// in:path
	Name string
	// The ID of the invalidation request
	// in:path
	ID string
}

const (
	ResourceNotFoundErrorCode         = 404
	InvalidationUnauthorizedErrorCode = 403
	InternalServerError               = 500
)

var (
	statusCodeReasons = map[int]string{
		InvalidationUnauthorizedErrorCode: "User is not entitled to invalidate distribution: %s",
		ResourceNotFoundErrorCode:         "Resource not found: %s",
	}
)

// swagger:model InvalidationError
type InvalidationError struct {
	InvalidationMeta
	Err  error `json:"-"`
	Code int   `json:"-"`
}

// swagger:model ErrorResponse
type ErrorResponse string

func (err InvalidationError) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return err.Status
}

func (err InvalidationError) Unwrap() error {
	return err.Err
}

func NewInvalidationError(code int, err error, args ...interface{}) error {
	return InvalidationError{
		Code: code,
		InvalidationMeta: InvalidationMeta{
			Status: fmt.Sprintf(statusCodeReasons[code], args...),
		},
		Err: err,
	}
}

func ErrorIsUnauthorized(err error) bool {
	var ierr InvalidationError
	if !errors.As(err, &ierr) {
		return false
	}

	return ierr.Code == InvalidationUnauthorizedErrorCode
}

func ErrorResourceNotFound(err error) bool {
	var ierr InvalidationError
	if !errors.As(err, &ierr) {
		return false
	}

	return ierr.Code == ResourceNotFoundErrorCode
}

//TODO this package will implement the api/v1beta1/DistributionService interface
