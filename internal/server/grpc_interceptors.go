package server

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type GrpcServerInterceptors struct {
	usersRepository   repository.UsersRepository
	secretsRepository repository.SecretsRepository
	keyManager        *crypto.KeyManager
}

// GrpcAuthentifiedInterceptor is the JwtAuth middleware combined with the Authentified interceptor for gRPC services.
func (i *GrpcServerInterceptors) GrpcAuthentifiedInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	tokenMetadata := md.Get("authorization")
	if len(tokenMetadata) == 0 {
		return nil, status.Error(codes.Unauthenticated, "you need to be authenticated to access this endpoint, no 'authorization' metadata found")
	} else if len(tokenMetadata) > 1 {
		return nil, status.Error(codes.InvalidArgument, "multiple 'authorization' metadata found, expected only one")
	}
	token := tokenMetadata[0]
	if len(token) > 7 && strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	if i.keyManager.GetState() != crypto.StateUnlocked {
		return nil, status.Error(codes.FailedPrecondition, "key manager is locked, please unlock it first")
	}

	parsedToken, err := jwt.ParseWithClaims(token, &service.JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		secret, err := i.secretsRepository.GetSecretByPath(ctx, i.keyManager, "internal/jwt_signing_key")
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "could not retrieve JWT signing key: %v", err)
		}

		return base64.StdEncoding.DecodeString(secret.DecryptedValue)
	})

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, status.Error(codes.Unauthenticated, "invalid token signature")
		}
		return nil, status.Errorf(codes.Unauthenticated, "failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := parsedToken.Claims.(*service.JwtClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid token claims type")
	}

	ctx = context.WithValue(ctx, "token", claims)

	return handler(ctx, req)
}
