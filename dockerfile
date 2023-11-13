# Етап 1: Збірка програми
FROM golang:latest as builder
WORKDIR /app
COPY go.mod ./
# go.sum не використовується, оскільки немає зовнішніх залежностей
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o harbor-bot .

# Етап 2: Створення кінцевого образу
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/harbor-bot .
CMD ["./harbor-bot"]
