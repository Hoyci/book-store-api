package utils

import (
	"context"
	"net/http"
	"reflect"

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

func GetClaimFromContext[T any](r *http.Request, key string) (T, bool) {
	var zeroValue T
	claimsCtx, ok := GetClaimsFromContext(r.Context())
	if !ok {
		return zeroValue, false
	}

	val := reflect.ValueOf(claimsCtx).Elem()
	field := val.FieldByName(key)

	if !field.IsValid() {
		return zeroValue, false
	}

	if value, ok := field.Interface().(T); ok {
		return value, true
	}

	return zeroValue, false
}
