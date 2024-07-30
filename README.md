# AI Dashboard REST API

## Overview

This is a RESTful API built with the Gin framework in Go. It provides a endpoint for generating posture-checks, policies and profiles.

## Features

- Fast and lightweight
- RESTful architecture
- Middleware support for logging and error handling

## Prerequisites

- Go 1.22.5

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/netzilo/aidashboard.git
   ```

2. Install the dependencies:
   ```bash
   go mod tidy
   ```

## Usage

To run the API, use the following command:

```bash
go run ./cmd/main.go
```

The API will start on `http://localhost:8080/api/ai-dashboard`.

To test via Swagger, go to `http://localhost:8080/swagger/index.html#/default/post_ai_dashboard`.

## API Endpoints

### AI-Dashboard

#### Create dashboard

- **POST** `/api/ai-dashboard`
- **Request Body**:
  ```json
  {
      "user_command": "Create a policy which allows SSH traffic from devops group to staging group. It should also enforce devops team to connect from a specific IP range 175.0.0.1 to 178.0.0.255.",
  }
  ```
- **Response**:
  - **201 Generated**: Dashboard generated successfully.
  - **400 Bad Request**: Invalid input.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - The web framework used.
