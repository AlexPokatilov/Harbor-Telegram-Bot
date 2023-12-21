# Stage 1
FROM golang:latest as builder
WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o harbor-telegram-bot .

# Stage 2
FROM alpine:latest as bot

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/harbor-telegram-bot .

ENV CHAT_ID=
ENV BOT_TOKEN=
ENV DEBUG_MODE=true

EXPOSE 441:441
CMD ["./harbor-telegram-bot"]
