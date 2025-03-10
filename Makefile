GO ?= GO111MODULE=on CGO_ENABLED=0 go

run:
	$(GO) run main.go

build:
	$(GO) build -o bootstrap main.go

tests:
	$(GO) test ./...

docker-run:
	docker build -t stubborn-notifier . && docker run -it stubborn-notifier

update-all:
	go get -u ./...
	go mod tidy

lint:
	gofumpt -w *.go
	golines --base-formatter=gofumpt --max-len=120 --no-reformat-tags -w .
	wsl --fix ./...
	golangci-lint run --fix
