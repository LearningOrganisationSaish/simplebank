FROM golang:1.23.2-alpine3.20 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY db/migrations ./db/migrations
COPY wait-for-it.sh .
COPY start.sh .
EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT ["/app/start.sh"]