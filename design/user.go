package design

import (
	. "goa.design/goa/v3/dsl"
)

var UserType = Type("User", func() {
	Attribute("id", Int, "Unique identifier for the user", func() {
		Example(1)
	})
	Attribute("username", String, "The username", func() {
		Example("alice")
		MinLength(3)
	})
	Attribute("created_at", String, "User creation timestamp", func() {
		Example("2025-06-30T12:00:00Z")
	})
	Attribute("updated_at", String, "User last update timestamp", func() {
		Example("2025-06-30T15:00:00Z")
	})

	Attribute("roles", ArrayOf(RoleType), "Roles assigned to the user")

	Required("id", "username", "created_at", "updated_at", "roles")
})

var _ = Service("users", func() {
	Description("User service manages user accounts and authentication")

	Method("create user", func() {
		Description("Create a new user")
		Payload(func() {
			Attribute("username", String, "Username of the new user", func() {
				Example("alice")
				MinLength(3)
			})
			Attribute("password", String, "Password (hashed or plain depending on implementation)", func() {
				Example("s3cr3t")
			})
			Required("username", "password")
		})
		Result(func() {
			Attribute("username", String, "The username of the created user", func() {
				Example("alice")
			})
			Attribute("id", Int, "Unique identifier for the user", func() {
				Example(2)
			})

			Required("id", "username")
		})
		Error("username_taken", ErrorResult, "Username already exists")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		HTTP(func() {
			POST("/users")
			Response(StatusCreated)
			Response("username_taken", StatusConflict)
			Response("invalid_parameters", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("list users", func() {
		ServerInterceptor(Authentified)

		Description("List all users")
		Result(ArrayOf(UserType))
		Error("internal_error", ErrorResult, "Internal server error")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			GET("/users")
			Response(StatusOK)
			Response("internal_error", StatusInternalServerError)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("delete user", func() {
		ServerInterceptor(IsAdmin)

		Description("Delete a user by username")
		Payload(func() {
			Attribute("username", String, "Username of the user to delete", func() {
				Example("alice")
				MinLength(3)
			})
			Required("username")
		})
		Error("user_not_found", ErrorResult, "User not found")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		Error("forbidden", ErrorResult, "Forbidden access")
		Error("unauthorized", ErrorResult, "Unauthorized access")
		HTTP(func() {
			DELETE("/users/{username}")
			Response(StatusOK)
			Response("user_not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
			Response("invalid_parameters", StatusBadRequest)
			Response("forbidden", StatusForbidden)
			Response("unauthorized", StatusUnauthorized)
		})
	})

	Method("auth user", func() {
		Description("Authenticate a user with username and password")
		Payload(func() {
			Attribute("username", String, "Username", func() {
				Example("alice")
			})
			Attribute("password", String, "Password", func() {
				Example("s3cr3t")
			})
			Required("username", "password")
		})
		Result(func() {
			Attribute("username", String, "The username of the authenticated user", func() {
				Example("alice")
			})
			Attribute("token", String, "JWT or session token", func() {
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
			})
		})
		Error("unauthorized", ErrorResult, "Invalid username or password")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", ErrorResult, "Internal server error")
		HTTP(func() {
			POST("/users/auth")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("invalid_parameters", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("get operator token", func() {
		ServerInterceptor(Authentified)

		Description("Retrieve a JWT token that doesn't expire for operator use, corresponding to your user")
		Result(func() {
			Attribute("username", String, "The username of the account corresponding to the token", func() {
				Example("alice")
			})
			Attribute("token", String, "JWT or session token", func() {
				Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
			})
		})
		Error("internal_error", ErrorResult, "Internal server error")
		HTTP(func() {
			POST("/users/operator-token")
			Response(StatusOK)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
