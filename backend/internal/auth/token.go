package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sonixxks/Mentor/backend/internal/user"
)

type Claims struct {
	jwt.RegisteredClaims
	Role user.Role `json:"role"`
}

type TokenManager struct {
	secret    []byte
	accessTTL time.Duration
	issuer    string
	now       func() time.Time
}

type Principal struct {
	UserID uuid.UUID
	Role   user.Role
}

func NewTokenManager(secret string, accessTTL time.Duration, issuer string) (*TokenManager, error) {
	if accessTTL <= 0 {
		return nil, fmt.Errorf("JWT_ACCESS_TTL must be > 0, got - %s", accessTTL)
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 bytes, got - %v", len(secret))
	}

	if issuer == "" {
		return nil, fmt.Errorf("JWT_ISSUER cannot be empty")
	}

	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL, issuer: issuer, now: time.Now}, nil
}

func (m *TokenManager) IssueAccessToken(userID uuid.UUID, role user.Role) (string, time.Time, error) {

	if !role.Valid() {
		return "", time.Time{}, ErrInvalidRole
	}

	claims := Claims{Role: role}
	claims.Subject = userID.String()
	claims.Issuer = m.issuer
	tokenManagerNow := m.now()
	claims.ExpiresAt = jwt.NewNumericDate(tokenManagerNow.Add(m.accessTTL))
	claims.IssuedAt = jwt.NewNumericDate(tokenManagerNow)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedJwt, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("cannot sign token, %w", err)
	}

	return signedJwt, claims.ExpiresAt.Time, nil
}

func (m *TokenManager) ParseAccessToken(token string) (Principal, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{},
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Principal{}, ErrTokenExpired
		}
		return Principal{}, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Principal{}, ErrInvalidToken
	}
	subjectUuid, err := uuid.Parse(claims.Subject)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	if !claims.Role.Valid() {
		return Principal{}, ErrInvalidToken
	}

	return Principal{UserID: subjectUuid, Role: claims.Role}, nil

}
