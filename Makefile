gen:
	goa gen github.com/Vidalee/FishyKeys/design

test:
	go test ./...

.PHONY: gen test