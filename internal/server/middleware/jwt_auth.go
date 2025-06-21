package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func JWTMiddleware(secretsRepository repository.SecretsRepository, keyManager *crypto.KeyManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenHeaderValue := r.Header.Get("Authorization")
			if tokenHeaderValue == "" {
				next.ServeHTTP(w, r)
				return
			}

			if keyManager.GetState() != crypto.StateUnlocked {
				// If the user still has a JWT token but the key manager is locked, don't cancel all requests.
				// Requests using the JWT token need to have the KeyUnlocked interceptor to ensure the request's context
				// is populated with the JwtClaims.
				next.ServeHTTP(w, r)
				return
			}

			if len(tokenHeaderValue) > 7 && strings.HasPrefix(tokenHeaderValue, "Bearer ") {
				tokenHeaderValue = tokenHeaderValue[7:]
			}

			token, err := jwt.ParseWithClaims(tokenHeaderValue, &service.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				decryptedSecret, err := secretsRepository.GetSecretByPath(r.Context(), keyManager, "internal/jwt_signing_key")
				if err != nil {
					return nil, fmt.Errorf("could not retrieve JWT signing key: %w", err)
				}

				return base64.StdEncoding.DecodeString(decryptedSecret.DecryptedValue)
			})

			if err != nil {
				if errors.Is(err, jwt.ErrSignatureInvalid) {
					http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				} else {
					http.Error(w, "Failed to parse token: "+err.Error(), http.StatusInternalServerError)
				}
				return
			}

			if !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "token", token.Claims.(*service.JwtClaims))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
