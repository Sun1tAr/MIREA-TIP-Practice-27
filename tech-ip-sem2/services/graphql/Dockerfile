FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
COPY tools.go ./

# Жёстко правим replace
RUN go mod edit -replace github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/shared=/shared

RUN go mod download

COPY . .

RUN go run github.com/99designs/gqlgen generate
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/graphql ./cmd/graphql

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder --chown=appuser:appuser /app/bin/graphql /app/graphql

USER appuser
EXPOSE 8090
CMD ["/app/graphql"]