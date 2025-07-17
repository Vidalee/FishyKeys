package design

import (
	. "goa.design/goa/v3/dsl"
)

var RoleType = Type("Role", func() {
	Attribute("id", Int, "Unique identifier for the role", func() {
		Example(1)
	})
	Attribute("name", String, "Name of the role", func() {
		Example("admin")
	})
	Attribute("color", String, "Color associated with the role", func() {
		Example("#FF5733")
	})
	Attribute("admin", Boolean, "Is this role an admin role?", func() {
		Default(false)
	})
	Attribute("created_at", String, "Role creation timestamp", func() {
		Example("2025-06-30T12:00:00Z")
	})
	Attribute("updated_at", String, "Role last update timestamp", func() {
		Example("2025-06-30T15:00:00Z")
	})

	Required("id", "name", "color", "admin", "created_at", "updated_at")
})

var _ = Service("roles", func() {
	Description("Roles service manages roles")

	Method("list roles", func() {
		ServerInterceptor(Authentified)

		Description("List all roles")
		Result(ArrayOf(RoleType))
		Error("internal_error", ErrorResult, "Internal server error")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			GET("/roles")
			Response(StatusOK)
			Response("internal_error", StatusInternalServerError)
			Response("unauthorized", StatusUnauthorized)
		})
	})
})
