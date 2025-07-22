package server

import (
	"context"
	"encoding/base64"
	"fmt"
	gensecretspb "github.com/Vidalee/FishyKeys/gen/grpc/secrets/pb"
	gensecretsserver "github.com/Vidalee/FishyKeys/gen/grpc/secrets/server"
	keysvvr "github.com/Vidalee/FishyKeys/gen/http/key_management/server"
	rolessvvr "github.com/Vidalee/FishyKeys/gen/http/roles/server"
	secretssvvr "github.com/Vidalee/FishyKeys/gen/http/secrets/server"
	userssvvr "github.com/Vidalee/FishyKeys/gen/http/users/server"
	"github.com/Vidalee/FishyKeys/gen/key_management"
	"github.com/Vidalee/FishyKeys/gen/roles"
	"github.com/Vidalee/FishyKeys/gen/secrets"
	"github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/internal/server/middleware"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"github.com/jackc/pgx/v5/pgxpool"
	goahttp "goa.design/goa/v3/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"net/http"
)

func NewServers(pool *pgxpool.Pool) (http.Handler, *grpc.Server) {
	keyManager := crypto.GetDefaultKeyManager()

	globalSettingsRepo := repository.NewGlobalSettingsRepository(pool)
	usersRepo := repository.NewUsersRepository(pool)
	rolesRepo := repository.NewRolesRepository(pool)
	userRolesRepo := repository.NewUserRolesRepository(pool)
	secretsRepo := repository.NewSecretsRepository(pool)
	secretsAccessRepository := repository.NewSecretsAccessRepository(pool)

	keyService := service.NewKeyManagementService(keyManager, globalSettingsRepo, usersRepo, rolesRepo, userRolesRepo, secretsRepo)
	usersService := service.NewUsersService(keyManager, usersRepo, globalSettingsRepo, secretsRepo)
	secretsService := service.NewSecretsService(keyManager, usersRepo, rolesRepo, userRolesRepo, globalSettingsRepo, secretsRepo, secretsAccessRepository)
	rolesService := service.NewRolesService(rolesRepo)

	keyManagementEndpoints := keymanagement.NewEndpoints(keyService)
	usersEndpoints := users.NewEndpoints(usersService, &ServerUsersInterceptors{
		userRolesRepository: userRolesRepo,
		rolesRepository:     rolesRepo,
	})
	secretsEndpoints := secrets.NewEndpoints(secretsService, &ServerSecretsInterceptors{
		userRolesRepository: userRolesRepo,
		rolesRepository:     rolesRepo,
	})
	rolesEndpoints := roles.NewEndpoints(rolesService, &ServerRolesInterceptors{
		userRolesRepository: userRolesRepo,
		rolesRepository:     rolesRepo,
	})

	mux := goahttp.NewMuxer()
	requestDecoder := goahttp.RequestDecoder
	responseEncoder := goahttp.ResponseEncoder

	keyManagementHandler := keysvvr.New(keyManagementEndpoints, mux, requestDecoder, responseEncoder, nil, nil)
	usersHandler := userssvvr.New(usersEndpoints, mux, requestDecoder, responseEncoder, nil, nil)
	secretsHandler := secretssvvr.New(secretsEndpoints, mux, requestDecoder, responseEncoder, nil, nil)
	rolesHandler := rolessvvr.New(rolesEndpoints, mux, requestDecoder, responseEncoder, nil, nil)

	mux.Use(middleware.JWTMiddleware(secretsRepo, keyManager))

	keysvvr.Mount(mux, keyManagementHandler)
	userssvvr.Mount(mux, usersHandler)
	secretssvvr.Mount(mux, secretsHandler)
	rolessvvr.Mount(mux, rolesHandler)

	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	gensecretspb.RegisterSecretsServer(grpcSrv, gensecretsserver.New(secretsEndpoints, nil))
	reflection.Register(grpcSrv)

	return mux
}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Printf("Handling %s\n", info.FullMethod)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		fmt.Println("Metadata received:")
		for key, values := range md {
			fmt.Printf("  %s: %v\n", key, values)
		}
	} else {
		fmt.Println("No metadata found in context")
	}

	return handler(ctx, req)
}
