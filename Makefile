GO ?= GO111MODULE=on CGO_ENABLED=0 go

run:
	$(GO) run ${TARGET}

run-main:
	$(GO) run entrypoints/app_entrypoint/main.go

build:
	$(GO) build -o bootstrap ${TARGET}

tests:
	$(GO) test ./...

docker-run:
	docker build --build-arg TARGET=${TARGET} -t telemetry-go . && docker run -it telemetry-go

update-all:
	go get -u ./...
	go mod tidy

lint:
	gofumpt -w *.go
	golines --base-formatter=gofumpt --max-len=120 --no-reformat-tags -w .
	wsl --fix ./...
	golangci-lint run --fix
