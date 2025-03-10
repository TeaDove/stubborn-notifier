# Start by building the application.
FROM golang:1.23-bullseye AS build

WORKDIR /src
COPY go.mod go.sum main.go ./

ENV CGO_ENABLED=0
RUN go build -o stub && rm stub
RUN go get ./...

COPY . .

ARG TARGET
RUN TARGET=$TARGET make build

## Now copy it into our base image.
FROM debian:trixie

RUN rm -rf /var/lib/apt/lists/* \
    && apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl ffmpeg
RUN update-ca-certificates
RUN rm -rf /var/lib/apt/lists/*

COPY --from=build /src/bootstrap /

CMD ["/bootstrap"]
