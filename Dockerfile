# Build the app in a multi-stage build
FROM --platform=linux/amd64 golang:1.22-alpine AS builder

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/

RUN go build -o main ./cmd/telegram-butler

# Switch to a smaller image for the final container
FROM alpine:3.19.1

COPY --from=builder /go/src/app/main .

CMD ["./main"]
