gen:
	goa gen github.com/Vidalee/FishyKeys/backend/design

test:
	go test ./...

.PHONY: gen test