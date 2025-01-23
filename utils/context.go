package utils

import (
	"context"

	"github.com/hoyci/book-store-api/types"
)

type ContextKey string

const ClaimsContextKey ContextKey = "claims"

func SetClaimsToContext(ctx context.Context, claims *types.CustomClaims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

func GetClaimsFromContext(ctx context.Context) (*types.CustomClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*types.CustomClaims)
	return claims, ok
}
