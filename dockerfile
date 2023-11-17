FROM golang:latest as builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o harbor-telegram-bot .

FROM alpine:latest as bot
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/harbor-telegram-bot .
CMD ["./harbor-telegram-bot"]
