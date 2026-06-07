FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]
