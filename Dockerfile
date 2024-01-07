FROM golang:latest

WORKDIR /app

COPY . .

EXPOSE 8000

CMD ["go", "run", "./cmd/api/main.go"]