// Package middleware содержит обёртки для обработчиков REST API.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// TokenParser — проверка access-токена (реализует auth.TokenManager).
type TokenParser interface {
	ParseAccessToken(token string) (auth.Principal, error)
}

type principalKey struct{}

// WithPrincipal кладёт пользователя в контекст (нужно и для тестов обработчиков).
func WithPrincipal(ctx context.Context, p auth.Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom достаёт пользователя, которого положил RequireAuth или OptionalAuth.
func PrincipalFrom(ctx context.Context) (auth.Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(auth.Principal)
	return p, ok
}

type Auth struct {
	tokens TokenParser
}

func NewAuth(tokens TokenParser) *Auth {
	return &Auth{tokens: tokens}
}

// Require пропускает только запросы с валидным Bearer-токеном.
// Если переданы роли, пользователь должен иметь одну из них (иначе 403).
func (a *Auth) Require(next http.HandlerFunc, roles ...user.Role) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			httpx.SendAPIError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "missing bearer token")
			return
		}

		principal, err := a.tokens.ParseAccessToken(token)
		if err != nil {
			msg := "invalid token"
			if errors.Is(err, auth.ErrTokenExpired) {
				msg = "token expired"
			}
			httpx.SendAPIError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, msg)
			return
		}

		if len(roles) > 0 && !slices.Contains(roles, principal.Role) {
			httpx.SendAPIError(w, http.StatusForbidden, httpx.CodeForbidden, "action is not allowed for this role")
			return
		}

		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
	})
}

// Optional — для публичных ручек, которые отвечают авторизованным иначе
// (например, каталог с процентом совместимости). Невалидный токен — 401,
// чтобы клиент понял, что пора обновить токен.
func (a *Auth) Optional(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bearerToken(r); !ok {
			next.ServeHTTP(w, r)
			return
		}
		a.Require(next).ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}
