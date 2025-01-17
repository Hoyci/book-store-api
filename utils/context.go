package utils

import "context"

type ContextKey string

const ClaimsContextKey ContextKey = "claims"

func SetClaimsToContext(ctx context.Context, claims *CustomClaims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

func GetClaimsFromContext(ctx context.Context) (CustomClaims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(CustomClaims)
	return claims, ok
}
