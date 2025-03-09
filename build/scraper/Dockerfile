# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.24 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o scraper ./cmd/scraper

# Deploy the application binary into a lean image
FROM alpine:latest

WORKDIR /

COPY --from=build-stage /app/scraper /scraper

ENTRYPOINT ["/scraper"]
