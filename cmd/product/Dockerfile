FROM golang:1.24-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/product
RUN CGO_ENABLED=0 GOOS=linux go build -o /product-service ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /product-service /usr/local/bin/product-service
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/product-service"]
