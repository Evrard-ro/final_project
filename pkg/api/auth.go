package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtSecret = "your-secret-key-change-in-production"

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// passwordHash создаёт SHA256 хеш пароля
func passwordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// createToken создаёт JWT токен с хешем пароля
func createToken(password string) (string, error) {
	hash := passwordHash(password)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"hash": hash,
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(jwtSecret))
}

// validateToken проверяет JWT токен и сравнивает с текущим паролем
func validateToken(tokenString string, currentPassword string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Проверяем срок действия
	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return false
	}

	// Сравниваем хеш из токена с хешем текущего пароля
	hash, ok := claims["hash"].(string)
	if !ok {
		return false
	}

	return hash == passwordHash(currentPassword)
}

// signInHandler обрабатывает POST /api/signin
func signInHandler(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, SignInResponse{Error: "Некорректный запрос"})
		return
	}

	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		writeJSON(w, SignInResponse{Error: "Аутентификация не настроена"})
		return
	}

	if req.Password != password {
		writeJSON(w, SignInResponse{Error: "Неверный пароль"})
		return
	}

	token, err := createToken(password)
	if err != nil {
		writeJSON(w, SignInResponse{Error: "Ошибка создания токена"})
		return
	}

	writeJSON(w, SignInResponse{Token: token})
}

// auth middleware для проверки аутентификации
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, установлен ли пароль
		password := os.Getenv("TODO_PASSWORD")
		if len(password) > 0 {
			var jwtToken string

			// Получаем куку token
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtToken = cookie.Value
			}

			// Валидируем токен
			valid := validateToken(jwtToken, password)

			if !valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	})
}
