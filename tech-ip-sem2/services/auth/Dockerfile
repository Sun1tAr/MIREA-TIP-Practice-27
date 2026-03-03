FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем всё, включая graph папку с резолверами
COPY . .

# Убеждаемся что go.mod и tools.go есть
COPY go.mod go.sum ./
COPY tools.go ./

# Жёстко правим replace
RUN go mod edit -replace github.com/sun1tar/MIREA-TIP-Practice-27/tech-ip-sem2/shared=/shared

# Качаем зависимости
RUN go mod download

# Генерируем GraphQL код (использует schema.graphqls и резолверы)
RUN go run github.com/99designs/gqlgen generate

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/graphql ./cmd/graphql

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder --chown=appuser:appuser /app/bin/graphql /app/graphql

# Копируем schema.graphqls для рефлексии (нужно для генерации)
COPY --from=builder /app/graph/schema.graphqls /app/graph/schema.graphqls

USER appuser
EXPOSE 8090
CMD ["/app/graphql"]