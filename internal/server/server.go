package server

import (
	"encoding/base64"
	keysvvr "github.com/Vidalee/FishyKeys/gen/http/key_management/server"
	userssvvr "github.com/Vidalee/FishyKeys/gen/http/users/server"
	"github.com/Vidalee/FishyKeys/gen/key_management"
	"github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/internal/server/middleware"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"github.com/jackc/pgx/v5/pgxpool"
	goahttp "goa.design/goa/v3/http"
	"net/http"
)

func NewServer(pool *pgxpool.Pool) http.Handler {
	keyManager := crypto.GetDefaultKeyManager()

	globalSettingsRepo := repository.NewGlobalSettingsRepository(pool)
	usersRepo := repository.NewUsersRepository(pool)
	rolesRepo := repository.NewRolesRepository(pool)
	userRolesRepo := repository.NewUserRolesRepository(pool)
	secretsRepo := repository.NewSecretsRepository(pool)

	keyService := service.NewKeyManagementService(keyManager, globalSettingsRepo, usersRepo, rolesRepo, userRolesRepo, secretsRepo)
	userService := service.NewUsersService(keyManager, usersRepo, globalSettingsRepo, secretsRepo)

	keyManagementEndpoints := keymanagement.NewEndpoints(keyService)
	usersEndpoints := users.NewEndpoints(userService, &ServerInterceptors{
		userRolesRepository: userRolesRepo,
		rolesRepository:     rolesRepo,
	})

	mux := goahttp.NewMuxer()
	requestDecoder := goahttp.RequestDecoder
	responseEncoder := goahttp.ResponseEncoder

	keyManagementHandler := keysvvr.New(keyManagementEndpoints, mux, requestDecoder, responseEncoder, nil, nil)
	usersHandler := userssvvr.New(usersEndpoints, mux, requestDecoder, responseEncoder, nil, nil)

	mux.Use(middleware.JWTMiddleware(secretsRepo, keyManager))

	keysvvr.Mount(mux, keyManagementHandler)
	userssvvr.Mount(mux, usersHandler)

	return mux
}
