FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go build -o api-server .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/api-server ./

EXPOSE 5000

CMD ["./api-server"]
