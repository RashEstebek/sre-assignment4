FROM golang:1.20-alpine

WORKDIR /app/product-service
COPY . .

RUN go mod download

EXPOSE 50052

CMD ["go", "run", "./cmd/main.go"]