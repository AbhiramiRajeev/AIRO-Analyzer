FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /airo-analyzer ./cmd/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /airo-analyzer .

EXPOSE 8080

CMD ["./airo-analyzer"]
