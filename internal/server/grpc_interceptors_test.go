package server

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	repositorymocks "github.com/Vidalee/FishyKeys/gen/mocks"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const jwtTestSecret = "jwt test secret"

func createTestToken(secret []byte, claims *service.JwtClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func TestGrpcAuthentifiedInterceptor_NoAuthHeader(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()

	interceptor := &GrpcServerInterceptors{
		secretsRepository: mockRepo,
		keyManager:        keyManager,
	}

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return nil, nil
	}

	ctx := context.Background()
	resp, err := interceptor.GrpcAuthentifiedInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.False(t, handlerCalled)
}

func TestGrpcAuthentifiedInterceptor_KeyManagerLocked(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()

	interceptor := &GrpcServerInterceptors{
		secretsRepository: mockRepo,
		keyManager:        keyManager,
	}

	md := metadata.Pairs("authorization", "Bearer token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.GrpcAuthentifiedInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called when key manager is locked")
		return nil, nil
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestGrpcAuthentifiedInterceptor_InvalidToken(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	_ = keyManager.SetNewMasterKey([]byte("masterkey"), 1, 1)

	interceptor := &GrpcServerInterceptors{
		secretsRepository: mockRepo,
		keyManager:        keyManager,
	}

	md := metadata.Pairs("authorization", "Bearer invalid.token.value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.GrpcAuthentifiedInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called for invalid token")
		return nil, nil
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "failed to parse token")
}

func TestGrpcAuthentifiedInterceptor_InvalidSignature(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	_ = keyManager.SetNewMasterKey([]byte("masterkey"), 1, 1)

	interceptor := &GrpcServerInterceptors{
		secretsRepository: mockRepo,
		keyManager:        keyManager,
	}

	mockRepo.On("GetSecretByPath", mock.Anything, mock.Anything, mock.Anything).
		Return(&repository.DecryptedSecret{
			DecryptedValue: base64.StdEncoding.EncodeToString([]byte(jwtTestSecret)),
		}, nil)

	// Use a wrong secret to generate token
	tokenString, _ := createTestToken([]byte(jwtTestSecret+"invalid"), &service.JwtClaims{Username: "user", UserID: 1})

	md := metadata.Pairs("authorization", "Bearer "+tokenString)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.GrpcAuthentifiedInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called for bad signature")
		return nil, nil
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "invalid token signature")
}

func TestGrpcAuthentifiedInterceptor_RepositoryError(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	_ = keyManager.SetNewMasterKey([]byte("masterkey"), 1, 1)

	interceptor := &GrpcServerInterceptors{
		secretsRepository: mockRepo,
		keyManager:        keyManager,
	}

	mockRepo.On("GetSecretByPath", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	tokenString, _ := createTestToken([]byte(jwtTestSecret), &service.JwtClaims{Username: "user", UserID: 1})

	md := metadata.Pairs("authorization", "Bearer "+tokenString)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.GrpcAuthentifiedInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called on repo error")
		return nil, nil
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "could not retrieve JWT signing key")
}

func TestGrpcAuthentifiedInterceptor_ValidToken(t *testing.T) {
	mockRepo := repositorymocks.NewMockSecretsRepository(t)
	keyManager := crypto.GetDefaultKeyManager()
	_ = keyManager.SetNewMasterKey([]byte("masterkey"), 1, 1)

	interceptor := &GrpcServerInterceptors{
		secretsRepository: mockRepo,
		keyManager:        keyManager,
	}

	mockRepo.On("GetSecretByPath", mock.Anything, mock.Anything, "internal/jwt_signing_key").
		Return(&repository.DecryptedSecret{
			DecryptedValue: base64.StdEncoding.EncodeToString([]byte(jwtTestSecret)),
		}, nil)

	tokenString, err := createTestToken([]byte(jwtTestSecret), &service.JwtClaims{
		Username: "valid_user",
		UserID:   1,
	})
	assert.NoError(t, err)

	md := metadata.Pairs("authorization", "Bearer "+tokenString)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.GrpcAuthentifiedInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req interface{}) (interface{}, error) {
		claims := ctx.Value("token").(*service.JwtClaims)
		assert.Equal(t, "valid_user", claims.Username)
		assert.Equal(t, 1, claims.UserID)
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
	mockRepo.AssertExpectations(t)
}
