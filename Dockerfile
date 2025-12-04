FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/main.go


FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["./server"]
