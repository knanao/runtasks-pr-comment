FROM golang:1.21.3-alpine3.18 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /server

FROM alpine:3.18
RUN apk --no-cache add ca-certificates

COPY --from=builder /server ./
RUN chmod +x ./server

EXPOSE 8080
ENTRYPOINT ["./server"]
