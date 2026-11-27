package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/user"
)

// PendingRegistration — данные регистрации до подтверждения почты.
// Живёт в RegistrationStore с TTL = CodeSettings.TTL. После verify-email
// превращается в строки users + profiles и удаляется.
type PendingRegistration struct {
	Email        string
	PasswordHash string
	Role         user.Role
	FirstName    string
	LastName     string
	MiddleName   *string
	ConsentAt    time.Time // момент terms_accepted, при подтверждении пишется в users.consent_at
	CodeHash     string    // HMAC-SHA256 от кода, сам код не хранится
	Attempts     int
}

// CodeSettings — правила одноразовых кодов. Заполняется в app.go из конфига.
type CodeSettings struct {
	TTL            time.Duration // срок жизни кода (api.yaml: 10 минут)
	ResendInterval time.Duration // не чаще раза в N (api.yaml: 60 секунд)
	MaxAttempts    int           // попыток ввода (api.yaml: 3)
	// HashSecret — ключ для HMAC кода. Без ключа 6-значный код по его SHA-256
	// перебирается за доли секунды, если кто-то прочитает Redis.
	HashSecret []byte
}

// hashCode возвращает HMAC-SHA256(secret, code) в hex.
// Сравнивать хэши при проверке кода — через hmac.Equal (постоянное время).
func hashCode(secret []byte, code string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

// RandomCodeGenerator — реализация CodeGenerator на crypto/rand.
type RandomCodeGenerator struct{}

func (RandomCodeGenerator) Generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// verifyEmailLimiterKey — ключ ResendLimiter для писем подтверждения почты.
// У сброса пароля и смены почты будут свои префиксы, чтобы лимиты не пересекались.
func verifyEmailLimiterKey(email string) string {
	return "verify-email:" + email
}
