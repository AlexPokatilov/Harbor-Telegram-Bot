# Stage 1
FROM golang:1.24 as builder
WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o harbor-telegram-bot .

# Stage 2
FROM alpine:3 as bot

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/harbor-telegram-bot /app/

ENV CHAT_ID=
ENV BOT_TOKEN=
ENV DEBUG=false

EXPOSE 441:441
CMD ["/app/harbor-telegram-bot"]
