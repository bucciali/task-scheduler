FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o todoapp .


FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/todoapp .
COPY web ./web

RUN apk add --no-cache sqlite

ENV TODO_PASSWORD=12345
ENV TODO_DBFILE=/app/pkg/db/scheduler.db
EXPOSE ${TODO_PORT}

EXPOSE 7540

CMD ["./todoapp"]
