package api

import (
	"net/http"

	"github.com/kanopy-platform/cdnvalidator/internal/server/middleware/authorization"
)

type APIHandler interface {
	RegisterRoutes(authz authorization.Authorizer, router *http.ServeMux)
}
