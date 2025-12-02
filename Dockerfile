FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main main.go

FROM alpine:3.17
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]
