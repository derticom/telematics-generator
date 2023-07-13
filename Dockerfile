FROM golang:latest 

RUN apt-get update && apt-get install -y netcat-openbsd

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/generator/

CMD ["/app/main"]
