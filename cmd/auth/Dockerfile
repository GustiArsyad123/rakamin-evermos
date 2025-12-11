FROM golang:1.24-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/auth
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /auth-service /usr/local/bin/auth-service
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/auth-service"]
