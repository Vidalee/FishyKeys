package design

import (
	. "goa.design/goa/v3/dsl"
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
})
