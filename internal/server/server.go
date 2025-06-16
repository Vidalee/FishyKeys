package server

import (
	"github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"net/http"

	keysvvr "github.com/Vidalee/FishyKeys/gen/http/key_management/server"
	userssvvr "github.com/Vidalee/FishyKeys/gen/http/users/server"
	"github.com/Vidalee/FishyKeys/gen/key_management"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
	goahttp "goa.design/goa/v3/http"
)

func NewServer(pool *pgxpool.Pool) http.Handler {
	keyManager := crypto.GetDefaultKeyManager()
	keyRepo := repository.NewGlobalSettingsRepository(pool)
	usersRepo := repository.NewUsersRepository(pool)
	rolesRepo := repository.NewRolesRepository(pool)
	userRolesRepo := repository.NewUserRolesRepository(pool)

	keyService := service.NewKeyManagementService(keyManager, keyRepo, usersRepo, rolesRepo, userRolesRepo)
	userService := service.NewUsersService(keyManager, usersRepo)

	keyManagementEndpoints := keymanagement.NewEndpoints(keyService)
	usersEndpoints := users.NewEndpoints(userService)

	mux := goahttp.NewMuxer()
	requestDecoder := goahttp.RequestDecoder
	responseEncoder := goahttp.ResponseEncoder

	keyManagementHandler := keysvvr.New(keyManagementEndpoints, mux, requestDecoder, responseEncoder, nil, nil)
	usersHandler := userssvvr.New(usersEndpoints, mux, requestDecoder, responseEncoder, nil, nil)

	keysvvr.Mount(mux, keyManagementHandler)
	userssvvr.Mount(mux, usersHandler)

	return mux
}
