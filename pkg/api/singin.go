package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	PassHash string `json:"pass_hash"` // хэш пароля
	jwt.RegisteredClaims
}

var Pass = os.Getenv("TODO_PASSWORD")

func CreateToken(password string, secretKey []byte) (string, error) {

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)), // токен на 8 часов
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		Password string `json:"password"`
	}
	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if Pass == "" || req.Password != Pass {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный пароль"})
		return
	}

	secretKey := []byte("my_key")
	tokenString, err := CreateToken(Pass, secretKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось создать токен"})
		return
	}
	err = json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	if err != nil {
		ErrorJson(w, http.StatusBadRequest, "problems with token")
	}

}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		if len(Pass) > 0 {
			var jwtPass string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtPass = cookie.Value
			}
			var valid bool
			// здесь код для валидации и проверки JWT-токена
			secretKey := []byte("my_key")

			// Парсим токен из строки jwt (кука "token")
			token, err := jwt.ParseWithClaims(jwtPass, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				// проверяем, что алгоритм подписи тот, что мы ожидаем
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secretKey, nil
			})

			if err == nil && token.Valid {
				// проверяем срок действия
				if claims, ok := token.Claims.(*Claims); ok {
					if claims.ExpiresAt != nil && claims.ExpiresAt.Time.After(time.Now()) {
						valid = true
					}
				}
			}

			if !valid {
				// возвращаем ошибку авторизации 401
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
