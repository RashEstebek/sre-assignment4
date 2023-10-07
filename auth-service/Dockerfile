FROM golang:1.20-alpine

WORKDIR /app/auth-service
COPY . .

RUN go mod download

EXPOSE 50051

CMD ["go", "run", "./cmd/main.go"]