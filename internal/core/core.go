package core

import "context"

type ClaimsKey string

const (
	ContextBoundaryKey ClaimsKey = "claims"
)

func GetClaims(ctx context.Context) []string {
	claims, ok := ctx.Value(ContextBoundaryKey).([]string)
	if !ok {
		return []string{}
	}
	return claims
}
