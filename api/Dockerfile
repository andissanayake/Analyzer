FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o api-server ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/api-server ./

EXPOSE 5000

CMD ["./api-server"]
