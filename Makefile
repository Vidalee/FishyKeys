gen:
    # Keep goa first since it removes every files the gen folder
	goa gen github.com/Vidalee/FishyKeys/design
	mockery

test:
	go test ./...

build:
	go build ./

.PHONY: gen test