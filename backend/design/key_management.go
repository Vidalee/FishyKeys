package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("fishykeys", func() {
	Description("The FishyKeys server handles master key operations")

	Method("create_master_key", func() {
		Description("Create a new master key and split it into shares")
		Payload(func() {
			Attribute("total_shares", Int, "Total number of shares to create", func() {
				Example(5)
			})
			Attribute("min_shares", Int, "Minimum number of shares required to reconstruct the key", func() {
				Example(3)
			})
			Required("total_shares", "min_shares")
		})
		Result(func() {
			Attribute("shares", ArrayOf(String), "The generated key shares")
		})
		Error("invalid_parameters", String, "Invalid parameters provided")
		Error("internal_error", String, "Internal server error")
		Error("key_already_exists", String, "A master key already exists")
		HTTP(func() {
			POST("/key_management/create_master_key")
			Response(StatusCreated)
			Response("invalid_parameters", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
		})
	})

	Method("get_key_status", func() {
		Description("Get the current status of the master key")
		Result(func() {
			Attribute("is_locked", Boolean, "Whether the key is currently locked")
			Attribute("current_shares", Int, "Number of shares currently held")
			Attribute("min_shares", Int, "Minimum number of shares required")
			Attribute("total_shares", Int, "Total number of shares")
			Required("is_locked", "current_shares", "min_shares", "total_shares")
		})
		Error("no_key_set", String, "No master key has been set")
		Error("internal_error", String, "Internal server error")
		HTTP(func() {
			GET("/key_management/status")
			Response(StatusOK)
			Response("no_key_set", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
		})
	})
})
