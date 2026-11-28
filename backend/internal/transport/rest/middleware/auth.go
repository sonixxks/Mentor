package middleware

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/platform/authctx"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/httpx"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

// TokenParser — проверка access-токена (реализует auth.TokenManager).
type TokenParser interface {
	ParseAccessToken(token string) (authctx.Principal, error)
}

// WithPrincipal кладёт пользователя в контекст (нужно и для тестов обработчиков).
func WithPrincipal(ctx context.Context, p authctx.Principal) context.Context {
	return authctx.WithPrincipal(ctx, p)
}

// PrincipalFrom достаёт пользователя, которого положил RequireAuth или OptionalAuth.
func PrincipalFrom(ctx context.Context) (authctx.Principal, bool) {
	return authctx.PrincipalFrom(ctx)
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
			httpx.SendError(w, http.StatusUnauthorized, httpx.ErrorResponse{ErrorCode: httpx.ErrUnauthorized, Message: "требуется авторизация"})
			return
		}

		principal, err := a.tokens.ParseAccessToken(token)
		if err != nil {
			msg := "недействительный токен"
			if errors.Is(err, auth.ErrTokenExpired) {
				msg = "срок действия токена истёк"
			}
			httpx.SendError(w, http.StatusUnauthorized, httpx.ErrorResponse{ErrorCode: httpx.ErrUnauthorized, Message: msg})
			return
		}

		if len(roles) > 0 && !slices.Contains(roles, principal.Role) {
			httpx.SendError(w, http.StatusForbidden, httpx.ErrorResponse{ErrorCode: httpx.ErrForbidden, Message: "действие недоступно для вашей роли"})
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
