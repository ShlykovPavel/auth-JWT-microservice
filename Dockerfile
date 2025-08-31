# Этап 1: Сборка приложения
FROM golang:alpine3.22 AS builder
WORKDIR /build_app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6 && \
    swag init -g cmd/auth-JWT-microservice/main.go --output docs --parseDependency --parseInternal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o auth-JWT-microservice ./cmd/auth-JWT-microservice

# Этап 2: Финальный образ
FROM alpine:latest AS final
RUN apk add --no-cache tzdata
RUN adduser -D -u 1000 appuser
WORKDIR /app
# Копируем бинарник и файлы конфигурации
COPY --from=builder /build_app/auth-JWT-microservice .
COPY --from=builder /build_app/config.yaml .
COPY --from=builder /build_app/internal/storage/database/migration ./internal/storage/database/migration
#COPY --from=builder /build_app/secret_config.yaml .
EXPOSE 8080
USER appuser
ENTRYPOINT ["./auth-JWT-microservice"]