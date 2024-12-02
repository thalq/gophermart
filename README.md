# GopherMart

![Go Version](https://img.shields.io/github/go-mod/go-version/thalq/gophermart)
![Build Status](https://github.com/thalq/gophermart/actions/workflows/statictest.yml/badge.svg)
![Build Status](https://github.com/thalq/gophermart/actions/workflows/gophermart.yml/badge.svg)
![Coverage Status](https://coveralls.io/repos/github/thalq/gopher_mart/badge.svg?branch=main)
![License](https://img.shields.io/github/license/thalq/gopher_mart)
![Go Report Card](https://goreportcard.com/badge/github.com/thalq/gopher_mart)

# Description
GopherMart is a loyalty points system that allows users to register, log in, and manage their orders and balance. The system interacts with an external accrual system to calculate loyalty points for each order.

## Features
- User registration and authentication
- Order management
- Balance management
- Withdrawal requests
- Interaction with an external accrual system

# Getting Started
## Prerequisites
- Go 1.22.5 or later
- PostgreSQL

## Installation
1. Clone the repository:
```Git
git clone https://github.com/thalq/gophermart.git
cd gophermart
```

2. Initialize the Go module:
```Go
go mod tidy
```

3. Configure environment variables:
```bash
export RUN_ADDRESS="localhost:8081"
export DATABASE_URI="postgres://postgres:postgres@localhost/gophermart?sslmode=disable"
export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8080"
```

## Running the Application
1. Build and run the application:
```Go
go build -o gophermart cmd/gophermart/main.go
./gophermart
```
2. The server will start on the address specified in the RUN_ADDRESS environment variable.

## API Endpoints
```Go
POST /api/user/register - Register a new user
POST /api/user/login - Authenticate a user
POST /api/user/orders - Upload a new order
GET /api/user/orders - Get the list of orders
GET /api/user/balance - Get the user's balance
POST /api/user/balance/withdraw - Request a withdrawal
GET /api/user/withdrawals - Get the list of withdrawals
```

## Configuration
The application can be configured using environment variables or command-line flags:
```
RUN_ADDRESS or -a - Address to run the server
DATABASE_URI or -d - Database connection URI
ACCRUAL_SYSTEM_ADDRESS or -r - Accrual system address
```

## Running Tests
...tests are still in development :)

