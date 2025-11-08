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

	Method("create role", func() {
		ServerInterceptor(IsAdmin)

		Description("Create a new role")
		Payload(func() {
			Attribute("name", String, "Name or the role", func() {
				Example("team_sre")
				MinLength(1)
			})
			Attribute("color", String, "Color of the role", func() {
				Example("#33FF57")
				Pattern("^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$")
			})
			Required("name", "color")
		})
		Result(func() {
			Attribute("name", String, "The name of the created role", func() {
				Example("team_sre")
			})
			Attribute("color", String, "The color of the created role", func() {
				Example("#33FF57")
			})
			Attribute("id", Int, "Unique identifier for the role", func() {
				Example(2)
			})
			Required("id", "name", "color")
		})
		Error("role_taken", ErrorResult, "Role name already exists")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		Error("forbidden", ErrorResult, "Forbidden access")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			POST("/roles")
			Response(StatusCreated)
			Response("role_taken", StatusConflict)
			Response("invalid_parameters", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
			Response("forbidden", StatusForbidden)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("delete role", func() {
		ServerInterceptor(IsAdmin)

		Description("Delete a role byd id")
		Payload(func() {
			Attribute("id", Int, "ID of the role to delete", func() {
				Example(2)
			})
			Required("id")
		})
		Error("role_not_found", ErrorResult, "Role not found")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		Error("forbidden", ErrorResult, "Forbidden access")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			DELETE("/roles/{id}")
			Response(StatusOK)
			Response("role_not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
			Response("invalid_parameters", StatusBadRequest)
			Response("forbidden", StatusForbidden)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("assign role to user", func() {
		ServerInterceptor(IsAdmin)

		Description("Assign a role to a user")
		Payload(func() {
			//user id and role id
			Attribute("user_id", Int, "ID of the user to assign the role to", func() {
				Example(2)
			})
			Attribute("role_id", Int, "ID of the role to assign to the user", func() {
				Example(1)
			})

			Required("user_id", "role_id")
		})

		Error("user_not_found", ErrorResult, "User not found")
		Error("role_not_found", ErrorResult, "Role not found")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		Error("forbidden", ErrorResult, "Forbidden access")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			POST("/roles/assign")
			Response(StatusOK)
			Response("user_not_found", StatusNotFound)
			Response("role_not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
			Response("invalid_parameters", StatusBadRequest)
			Response("forbidden", StatusForbidden)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("unassign role to user", func() {
		ServerInterceptor(IsAdmin)

		Description("Unassign a role to a user")
		Payload(func() {
			Attribute("user_id", Int, "ID of the user to unassign the role to", func() {
				Example(2)
			})
			Attribute("role_id", Int, "ID of the role to unassign to the user", func() {
				Example(1)
			})

			Required("user_id", "role_id")
		})

		Error("user_not_found", ErrorResult, "User not found")
		Error("role_not_found", ErrorResult, "Role not found")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		Error("forbidden", ErrorResult, "Forbidden access")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			POST("/roles/unassign")
			Response(StatusOK)
			Response("user_not_found", StatusNotFound)
			Response("role_not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
			Response("invalid_parameters", StatusBadRequest)
			Response("forbidden", StatusForbidden)
			Response("unauthorized", StatusUnauthorized)
		})
	})
})
