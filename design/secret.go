package design

import (
	. "goa.design/goa/v3/dsl"
)

var SecretInfoType = Type("SecretInfo", func() {
	Attribute("path", String, "The original path of the secret", func() {
		Example("customers/google/api_key")
	})

	Attribute("owner", UserType, "The owner of the secret")
	Attribute("authorized_users", ArrayOf(UserType), "Members authorized to access the secret")
	Attribute("authorized_roles", ArrayOf(RoleType), "Roles authorized to access the secret")
	Attribute("created_at", String, "Creation timestamp of the secret", func() {
		Example("2025-06-30T12:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp of the secret", func() {
		Example("2025-06-30T15:00:00Z")
	})

	Required("path", "owner", "authorized_users", "authorized_roles", "created_at", "updated_at")
})

var SecretInfoSummaryType = Type("SecretInfoSummary", func() {
	Attribute("path", String, "The original path of the secret", func() {
		Example("customers/google/api_key")
	})

	Attribute("owner", UserType, "The owner of the secret")
	Attribute("created_at", String, "Creation timestamp of the secret", func() {
		Example("2025-06-30T12:00:00Z")
	})
	Attribute("updated_at", String, "Last update timestamp of the secret", func() {
		Example("2025-06-30T15:00:00Z")
	})

	Attribute("users", ArrayOf(UserType), "Users authorized to access the secret")
	Attribute("roles", ArrayOf(RoleType), "Roles authorized to access the secret")

	Required("path", "owner", "created_at", "updated_at", "users", "roles")
})

var _ = Service("secrets", func() {
	Description("User service manages user accounts and authentication")

	Error("invalid_parameters", ErrorResult, "Invalid token path")
	Error("unauthorized", ErrorResult, "Unauthorized access")
	Error("forbidden", ErrorResult, "Forbidden access")
	Error("internal_error", ErrorResult, "Internal server error")

	Method("list secrets", func() {
		ServerInterceptor(Authentified)

		Description("Retrieve all secrets you have access to")
		Result(ArrayOf(SecretInfoSummaryType), "List of secrets you have access to")
		Error("secret_not_found", ErrorResult, "Secret not found")
		HTTP(func() {
			GET("/secrets")
			Response(StatusOK)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

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
		Error("secret_not_found", ErrorResult, "Secret not found")
		HTTP(func() {
			GET("/secrets/{path}/value")
			Response(StatusOK)
			Response("secret_not_found", StatusNotFound)
			Response("invalid_parameters", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("operator get secret value", func() {
		Description("Retrieve a secret value using GRPC")
		Payload(func() {
			Field(1, "path", String, "Base64 encoded secret's path", func() {
				Example("L2N1c3RvbWVycy9nb29nbGUvYXBpX2tleQ==")
				MinLength(2)
			})
			Required("path")
		})
		Result(func() {
			Field(1, "value", String, "The secret value", func() {
				Example("SECRET_API_KEY")
			})
			Field(2, "path", String, "The original path of the secret", func() {
				Example("customers/google/api_key")
			})
		})
		Error("secret_not_found", ErrorResult, "Secret not found")

		GRPC(func() {
			Response(CodeOK)
			Response("secret_not_found", CodeNotFound)
			Response("invalid_parameters", CodeInvalidArgument)
			Response("unauthorized", CodeUnauthenticated)
			Response("forbidden", CodePermissionDenied)
			Response("internal_error", CodeInternal)
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
		Error("secret_not_found", ErrorResult, "Secret not found")
		HTTP(func() {
			GET("/secrets/{path}")
			Response(StatusOK)
			Response("secret_not_found", StatusNotFound)
			Response("invalid_parameters", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("create secret", func() {
		ServerInterceptor(Authentified)

		Description("Create a secret")
		Payload(func() {
			Attribute("path", String, "Base64 encoded secret's path", func() {
				Example("L2N1c3RvbWVycy9nb29nbGUvYXBpX2tleQ==")
				MinLength(2)
			})
			Attribute("value", String, "The secret value", func() {
				Example("SECRET_API_KEY123")
				MinLength(1)
			})
			Attribute("authorized_users", ArrayOf(Int), "Users IDs authorized to access the secret", func() {
				Example([]int{1, 2, 3})
			})
			Attribute("authorized_roles", ArrayOf(Int), "Role IDs authorized to access the secret", func() {
				Example([]int{1, 2})
			})
			Required("path", "value", "authorized_users", "authorized_roles")
		})
		HTTP(func() {
			POST("/secrets")
			Response(StatusCreated)
			Response("invalid_parameters", StatusBadRequest)
			Response("unauthorized", StatusUnauthorized)
			Response("forbidden", StatusForbidden)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
