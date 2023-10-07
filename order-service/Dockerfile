FROM golang:1.20-alpine

WORKDIR /app/order-service
COPY . .

RUN go mod download

EXPOSE 50053

CMD ["go", "run", "./cmd/main.go"]