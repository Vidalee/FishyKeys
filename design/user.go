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
			})
			Attribute("password", String, "Password (hashed or plain depending on implementation)", func() {
				Example("s3cr3t")
			})
			Required("username", "password")
		})
		Result(func() {
			Attribute("username", String, "The username of the created user")
		})
		Error("username_taken", String, "Username already exists")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			POST("/users")
			Response(StatusCreated)
			Response("username_taken", StatusConflict)
			Response("invalid_parameters", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("list users", func() {
		Description("List all users")
		Result(ArrayOf(UserType))
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			GET("/users")
			Response(StatusOK)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("delete user", func() {
		Description("Delete a user by username")
		Payload(func() {
			Attribute("username", String, "Username of the user to delete", func() {
				Example("alice")
			})
			Required("username")
		})
		Error("user_not_found", String, "User does not exist")
		Error("invalid_parameters", ErrorResult, "Invalid input")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			DELETE("/users/{username}")
			Response(StatusOK)
			Response("user_not_found", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
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
		Error("unauthorized", String, "Invalid username or password")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			POST("/users/auth")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
