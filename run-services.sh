#!/bin/bash

# Script to run all microservices in parallel
echo "Starting all microservices..."

# Run auth service (includes category routes)
echo "Starting auth service..."
go run ./cmd/auth/main.go &
AUTH_PID=$!

# Run product service
echo "Starting product service..."
go run ./cmd/product/main.go &
PRODUCT_PID=$!

# Run transaction service
echo "Starting transaction service..."
go run ./cmd/transaction/main.go &
TRANSACTION_PID=$!

# Run address service
echo "Starting address service..."
go run ./cmd/address/main.go &
ADDRESS_PID=$!

# Run store service
echo "Starting store service..."
go run ./cmd/store/main.go &
STORE_PID=$!

# Run file service
echo "Starting file service..."
go run ./cmd/file/main.go &
FILE_PID=$!

echo "All services started!"
echo "Auth service (port 8080): PID $AUTH_PID"
echo "Product service (port 8081): PID $PRODUCT_PID"
echo "Transaction service (port 8082): PID $TRANSACTION_PID"
echo "Address service (port 8083): PID $ADDRESS_PID"
echo "Store service (port 8084): PID $STORE_PID"
echo "File service (port 8085): PID $FILE_PID"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for all processes
wait