# Stage 1: Сборка бинарника
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY cmd/ ./cmd/
COPY ent/ ./ent/

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Stage 2: Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем только бинарник из builder
COPY --from=builder /build/server .

# Копируем статические файлы
COPY templates/ ./templates/
COPY static/ ./static/

# Устанавливаем права на выполнение
RUN chmod +x ./server

# Открываем порт
EXPOSE 9090

# Запускаем приложение
CMD ["./server"]
