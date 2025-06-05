package design

import (
	. "goa.design/goa/v3/dsl"
	cors "goa.design/plugins/v3/cors/dsl"
)

var _ = API("fishykeys", func() {
	Title("FishyKeys API")
	Description("The FishyKeys API for key management")
	Version("1.0")
	Server("fishykeys", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})

	cors.Origin("http://localhost:5173", func() {
		cors.Methods("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS")
		cors.Headers("Content-Type")
		cors.MaxAge(3600)
	})
})
