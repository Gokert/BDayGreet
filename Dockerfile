FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY . .

RUN go build -o app ./cmd/main.go

CMD ./app