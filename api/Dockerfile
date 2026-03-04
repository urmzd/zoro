FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git chromium
RUN go install github.com/swaggo/swag/v2/cmd/swag@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN swag init -g cmd/zoro/main.go -o docs --parseDependency --parseInternal
RUN CGO_ENABLED=0 go build -o /zoro ./cmd/zoro

FROM alpine:3.20

RUN apk add --no-cache ca-certificates chromium

COPY --from=builder /zoro /zoro

ENV CHROME_PATH=/usr/bin/chromium-browser

EXPOSE 8080
CMD ["/zoro"]
