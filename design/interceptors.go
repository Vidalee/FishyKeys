package design

import (
	. "goa.design/goa/v3/dsl"
)

var Authentified = Interceptor("Authentified", func() {
	Description("Server-side interceptor that validates JWT token and tenant ID")
})

var IsAdmin = Interceptor("IsAdmin", func() {
	Description("Server-side interceptor that checks if the user has admin privileges")
})
