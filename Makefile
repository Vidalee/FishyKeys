OS := $(shell uname 2>/dev/null || echo Windows)

ifeq ($(OS),Windows)
    COPY = xcopy /E /I /Y
    SEP = \\
else
    COPY = cp -r
    SEP = /
endif

gen:
    # Keep goa first since it removes every files the gen folder
	goa gen github.com/Vidalee/FishyKeys/design
	mockery
	$(COPY) gen$(SEP)grpc$(SEP)secrets$(SEP)pb operator$(SEP)gen

test:
	go test ./...

build:
	go build ./

.PHONY: gen test