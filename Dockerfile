FROM golang:1.17-alpine

WORKDIR /app

COPY . .

RUN go build -o paxos main.go

ENTRYPOINT ["/app/paxos"]