FROM golang:1.24.5-alpine AS builder
WORKDIR /app/cmd/server
COPY go.mod go.sum /app/
RUN cd /app && go mod download
COPY . /app/
RUN go build -o /app/server main.go

FROM alpine:3.22

WORKDIR /root/

COPY --from=builder /app/server .

CMD ["./server"]
