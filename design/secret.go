package design

import (
	. "goa.design/goa/v3/dsl"
)

var RoleType = Type("RoleType", func() {
	Attribute("id", Int, "Unique identifier for the role", func() {
		Example(1)
	})
	Attribute("name", String, "Name of the role", func() {
		Example("admin")
	})
})

var SecretInfoType = Type("SecretInfo", func() {
	Attribute("path", String, "The original path of the secret", func() {
		Example("customers/google/api_key")
	})

	Attribute("owner", UserType, "The owner of the secret")
	Attribute("authorized_members", ArrayOf(UserType), "Members authorized to access the secret")
	Attribute("authorized_roles", ArrayOf(RoleType), "Roles authorized to access the secret")
	Attribute("created_at", String, "Creation timestamp of the secret", func() {
		Example("2025-06-30T12:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp of the secret", func() {
		Example("2025-06-30T15:00:00Z")
	})
})

var _ = Service("secrets", func() {
	Description("User service manages user accounts and authentication")

	Error("secret_not_found", ErrorResult, "Token not found")
	Error("invalid_parameters", ErrorResult, "Invalid token path")
	Error("unauthorized", ErrorResult, "Unauthorized access")
	Error("internal_error", String, "Internal server error")

	Method("get secret value", func() {
		ServerInterceptor(Authentified)

		Description("Retrieve a secret value")
		Payload(func() {
			Attribute("path", String, "Base64 encoded secret's path", func() {
				Example("L2N1c3RvbWVycy9nb29nbGUvYXBpX2tleQ==")
				MinLength(2)
			})
			Required("path")
		})
		Result(func() {
			Attribute("value", String, "The secret value", func() {
				Example("SECRET_API_KEY")
			})
			Attribute("path", String, "The original path of the secret", func() {
				Example("customers/google/api_key")
			})
		})
		HTTP(func() {
			GET("/secrets/${path}/value")
			Response(StatusOK)
			Response("secret_not_found", StatusNotFound)
			Response("invalid_parameters", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("get secret", func() {
		ServerInterceptor(Authentified)

		Description("Retrieve a secret's information")
		Payload(func() {
			Attribute("path", String, "Base64 encoded secret's path", func() {
				Example("L2N1c3RvbWVycy9nb29nbGUvYXBpX2tleQ==")
				MinLength(2)
			})
			Required("path")
		})
		Result(SecretInfoType, "The secret's information")
		HTTP(func() {
			GET("/secrets/${path}")
			Response(StatusOK)
			Response("secret_not_found", StatusNotFound)
			Response("invalid_parameters", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
