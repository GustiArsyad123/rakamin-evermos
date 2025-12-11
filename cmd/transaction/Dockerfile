FROM golang:1.24-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/transaction
RUN CGO_ENABLED=0 GOOS=linux go build -o /transaction-service ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /transaction-service /usr/local/bin/transaction-service
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/transaction-service"]
