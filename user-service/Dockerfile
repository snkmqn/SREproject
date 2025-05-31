# === Stage 1: Build the Go app ===
FROM golang:1.24 AS builder

# Создаем рабочую директорию
WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./user-service/main.go

FROM alpine:latest

# Копируем только бинарник
COPY --from=builder /server /server

# Открываем порты: gRPC и метрики
EXPOSE 50051 8080

# Запускаем
ENTRYPOINT ["/server"]
