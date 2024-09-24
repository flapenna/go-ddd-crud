FROM golang:1.22.3

# Install jq to setup the healthcheck
RUN apt-get update && apt-get install -y jq

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/bin/go-ddd-crud ./cmd/server

CMD ["/app/bin/go-ddd-crud"]