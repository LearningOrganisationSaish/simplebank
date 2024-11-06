FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz


FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY --from=builder /app/migrate ./migrate
COPY db/migrations ./migrations
COPY wait-for-it.sh .
COPY start.sh .
EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]