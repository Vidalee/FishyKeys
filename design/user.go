package design

import (
	. "goa.design/goa/v3/dsl"
)

var UserType = Type("User", func() {
	Attribute("username", String, "The username")
	Attribute("created_at", String, "User creation timestamp")
	Attribute("updated_at", String, "User last update timestamp")
	Required("username", "created_at", "updated_at")
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
			Attribute("username", String, "The username of the created user")
			Attribute("id", Int, "Unique identifier for the user")
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
			Attribute("username", String, "The username of the authenticated user")
			Attribute("token", String, "JWT or session token")
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
})
