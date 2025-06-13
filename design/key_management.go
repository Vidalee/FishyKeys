package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = Service("key_management", func() {
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
			Response("key_already_exists", StatusConflict)
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

	Method("add_share", func() {
		Description("Add a share to unlock the master key")
		Payload(func() {
			Attribute("share", String, "One of the shares need to unlock the master key", func() {
				Example("EXAMPLEA5ZKwDn8Zotr3B+d+F+UzrcJ1Yhl2rU0")
			})
			Required("share")
		})
		Result(func() {
			Attribute("index", Int, "The index of the share added")
			Attribute("unlocked", Boolean, "Whether the master key has been unlocked")
			Required("index", "unlocked")
		})
		Error("invalid_parameters", String, "Invalid parameters provided")
		Error("internal_error", String, "Internal server error")
		Error("too_many_shares", String, "The maximum number of shares has been reached")
		Error("could_not_recombine", String, "Could not recombine the shares to unlock the key")
		Error("wrong_shares", String, "The key recombined from the shares is not the correct key")
		Error("no_key_set", String, "No master key has been set")
		Error("key_already_unlocked", String, "The master key is already unlocked")
		HTTP(func() {
			POST("/key_management/share")
			Response(StatusCreated)
			Response("invalid_parameters", StatusBadRequest)
			Response("internal_error", StatusInternalServerError)
			Response("too_many_shares", StatusConflict)
			Response("could_not_recombine", StatusBadRequest)
			Response("wrong_shares", StatusBadRequest)
			Response("no_key_set", StatusNotFound)
			Response("key_already_unlocked", StatusConflict)
		})
	})

	Method("delete_share", func() {
		Description("Delete a share from the key management system")
		Payload(func() {
			Attribute("index", Int, "The index of the share to delete", func() {
				Example(1)
			})
			Required("index")
		})
		Error("no_key_set", String, "No master key has been set")
		Error("internal_error", String, "Internal server error")
		Error("key_already_unlocked", String, "The master key is already unlocked")
		Error("wrong_index", String, "The index provided does not match any share")
		HTTP(func() {
			DELETE("/key_management/share")
			Response(StatusOK)
			Response("no_key_set", StatusNotFound)
			Response("internal_error", StatusInternalServerError)
			Response("key_already_unlocked", StatusConflict)
			Response("wrong_index", StatusBadRequest)
		})
	})
})
