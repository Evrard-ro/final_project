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

var (
	configPassword string
	authEnabled    bool
)

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// InitAuth инициализирует конфигурацию аутентификации
func InitAuth() {
	configPassword = os.Getenv("TODO_PASSWORD")
	authEnabled = len(configPassword) > 0
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignInRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SignInResponse{Error: "Некорректный запрос"})
		return
	}

	if !authEnabled {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SignInResponse{Error: "Аутентификация не настроена"})
		return
	}

	if req.Password != configPassword {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(SignInResponse{Error: "Неверный пароль"})
		return
	}

	token, err := createToken(configPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SignInResponse{Error: "Ошибка создания токена"})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SignInResponse{Token: token})
}

// auth middleware для проверки аутентификации
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, включена ли аутентификация
		if authEnabled {
			var jwtToken string

			// Получаем куку token
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtToken = cookie.Value
			}

			// Валидируем токен
			valid := validateToken(jwtToken, configPassword)

			if !valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	})
}
