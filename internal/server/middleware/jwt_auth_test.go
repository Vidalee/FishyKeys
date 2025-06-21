package middleware

import (
	"encoding/base64"
	"errors"
	"github.com/Vidalee/FishyKeys/gen/mocks"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const jwtTestToken = "test jwt secret"

func TestJWTMiddleware_NoAuthorizationHeader(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	assert.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 2)
	assert.NoError(t, err, "failed to set new master key")

	mw := JWTMiddleware(mockRepo, keyManager)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No token in context should be present
		token := r.Context().Value("token")
		assert.Nil(t, token)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	mw(testHandler).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestJWTMiddleware_KeyManagerNotUnlocked(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()

	mw := JWTMiddleware(mockRepo, keyManager)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No token in context should be present
		token := r.Context().Value("token")
		assert.Nil(t, token)
		w.WriteHeader(http.StatusOK)
	})

	mw(testHandler).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	assert.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 2)
	assert.NoError(t, err, "failed to set new master key")

	mw := JWTMiddleware(mockRepo, keyManager)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for invalid token")
	})

	mw(testHandler).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, resp.Body.String(), "Failed to parse token")
}

func TestJWTMiddleware_InvalidTokenSignature(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	assert.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 2)
	assert.NoError(t, err, "failed to set new master key")

	mockRepo.On("GetSecretByPath", mock.Anything, mock.Anything, mock.Anything).Return(
		&repository.DecryptedSecret{
			DecryptedValue: base64.StdEncoding.EncodeToString([]byte(jwtTestToken)),
		}, nil)

	secret := []byte(jwtTestToken + "invalid")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &service.JwtClaims{
		Username: "test_user",
	})

	assert.NoError(t, err)

	tokenString, err := token.SignedString(secret)
	mw := JWTMiddleware(mockRepo, keyManager)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for invalid token")
	})

	mw(testHandler).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Invalid token signature")
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	assert.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 2)
	assert.NoError(t, err, "failed to set new master key")

	mockRepo.On("GetSecretByPath", mock.Anything, mock.Anything, mock.Anything).Return(
		&repository.DecryptedSecret{
			DecryptedValue: base64.StdEncoding.EncodeToString([]byte(jwtTestToken)),
		}, nil)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &service.JwtClaims{
		Username: "test_user",
	})

	secret := []byte(jwtTestToken)
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	mw := JWTMiddleware(mockRepo, keyManager)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value("token").(*service.JwtClaims)
		assert.True(t, ok)
		assert.NotNil(t, claims)
		assert.Equal(t, "test_user", claims.Username)
		w.WriteHeader(http.StatusOK)
	})

	mw(testHandler).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	mockRepo.AssertExpectations(t)
}

func TestJWTMiddleware_RepositoryError(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	assert.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 2)
	assert.NoError(t, err, "failed to set new master key")

	mockRepo.On("GetSecretByPath", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	mw := JWTMiddleware(mockRepo, keyManager)

	secret := []byte(jwtTestToken)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &service.JwtClaims{
		Username: "test_user",
	})

	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on repository error")
	})

	mw(testHandler).ServeHTTP(resp, req)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.Contains(t, resp.Body.String(), "could not retrieve JWT signing key")
	mockRepo.AssertExpectations(t)
}
