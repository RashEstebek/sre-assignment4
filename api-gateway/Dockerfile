FROM golang:1.20-alpine

WORKDIR /app/api-gateway
COPY . .

RUN go mod download

EXPOSE 3001

CMD ["go", "run", "./cmd/main.go"]