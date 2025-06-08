package server

import (
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"net/http"

	"github.com/Vidalee/FishyKeys/gen/fishykeys"
	fishykeyssvr "github.com/Vidalee/FishyKeys/gen/http/fishykeys/server"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/jackc/pgx/v5/pgxpool"
	goahttp "goa.design/goa/v3/http"
)

// NewServer creates and configures the Goa server
func NewServer(pool *pgxpool.Pool) http.Handler {
	// Initialize dependencies
	keyManager := crypto.GetKeyManager()
	keyRepo := repository.NewGlobalSettingsRepository(pool)
	keyService := service.NewKeyManagementService(keyManager, keyRepo)

	// Create the service endpoints
	endpoints := fishykeys.NewEndpoints(keyService)

	// Set up the HTTP multiplexer and transport layer
	mux := goahttp.NewMuxer()
	requestDecoder := goahttp.RequestDecoder
	responseEncoder := goahttp.ResponseEncoder
	handler := fishykeyssvr.New(endpoints, mux, requestDecoder, responseEncoder, nil, nil)

	fishykeyssvr.Mount(mux, handler)

	return mux
}
