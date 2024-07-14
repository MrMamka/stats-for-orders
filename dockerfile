FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o stats-service cmd/main.go

CMD ["./stats-service", "-port", "8080"]
