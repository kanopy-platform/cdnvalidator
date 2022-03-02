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
	Status string `json:"status"`
}

type InvalidationResponse struct {
	InvalidationMeta
	Created time.Duration `json:"createTime"`
	Paths   []string      `json:"paths,omitempty"`
}

type InvalidationRequest struct {
	Paths []string `json:"paths"`
}

const (
	DistributionNotFoundErrorCode     = 404
	InvalidationUnAuthorizedErrorCode = 403
)

var (
	statusCodeReasons = map[int]string{
		InvalidationUnAuthorizedErrorCode: "User is not entitled to invalidate path: %s",
		DistributionNotFoundErrorCode:     "Distribution not found",
	}
)

type InvalidationError struct {
	InvalidationMeta
	Err  error `json:"-"`
	Code int   `json:"-"`
}

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

	return ierr.Code == 403
}

func ErrorDistributionNotFound(err error) bool {
	var ierr InvalidationError
	if !errors.As(err, &ierr) {
		return false
	}

	return ierr.Code == 404
}

//TODO this package will implement the api/v1beta1/DistributionService interface
