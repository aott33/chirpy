# Chirpy

A Twitter-like social platform REST API built with Go and PostgreSQL.

## Overview

Chirpy is a social media API that allows users to create accounts, post short messages (chirps), and manage their profiles. The API includes JWT-based authentication, refresh token management, and premium user upgrade functionality via webhooks.

## Prerequisites

- Go 1.25.4 or higher
- PostgreSQL database
- goose (for database migrations)
- sqlc (for generating database code)

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Create a `.env` file with the following variables:
```
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
JWT_SECRET=your-secret-key-here
POLKA_KEY=your-polka-api-key-here
```

3. Run database migrations:
```bash
goose postgres $DB_URL up
```

4. Build and run the server:
```bash
go build -o chirpy
./chirpy
```

The server will start on `http://localhost:8080`

## API Routes

### Health Check

**GET /api/healthz**
- Description: Server health check
- Authentication: None
- Response: 200 OK with "OK" text

### Chirps

**POST /api/chirps**
- Description: Create a new chirp
- Authentication: Required (JWT Bearer token)
- Request Body:
  ```json
  {
    "body": "string (max 140 characters)"
  }
  ```
- Response: 201 Created with chirp object

**GET /api/chirps**
- Description: Get all chirps
- Authentication: None
- Query Parameters:
  - `author_id` (optional): UUID - Filter chirps by author
  - `sort` (optional): String - Sort order (`asc` or `desc`, default is `asc`)
- Response: 200 OK with array of chirp objects

**GET /api/chirps/{chirpID}**
- Description: Get a specific chirp by ID
- Authentication: None
- Response: 200 OK with chirp object

**DELETE /api/chirps/{chirpID}**
- Description: Delete a chirp (must be owner)
- Authentication: Required (JWT Bearer token)
- Response: 204 No Content

### Users

**POST /api/users**
- Description: Create a new user account
- Authentication: None
- Request Body:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- Response: 201 Created with user object

**PUT /api/users**
- Description: Update user email and password
- Authentication: Required (JWT Bearer token)
- Request Body:
  ```json
  {
    "email": "newemail@example.com",
    "password": "newpassword123"
  }
  ```
- Response: 200 OK with updated user object

**POST /api/login**
- Description: Login and receive access tokens
- Authentication: None
- Request Body:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- Response: 200 OK with user object, JWT token, and refresh token

**POST /api/refresh**
- Description: Refresh JWT access token
- Authentication: Required (Bearer refresh token)
- Response: 200 OK with new JWT token

**POST /api/revoke**
- Description: Revoke a refresh token
- Authentication: Required (Bearer refresh token)
- Response: 204 No Content

### Webhooks

**POST /api/polka/webhooks**
- Description: Handle user upgrade events from Polka service
- Authentication: Required (ApiKey header)
- Request Body:
  ```json
  {
    "event": "user.upgraded",
    "data": {
      "user_id": "uuid-string"
    }
  }
  ```
- Response: 204 No Content

### Admin

**GET /admin/metrics**
- Description: View file server hit metrics
- Authentication: None
- Response: 200 OK with HTML page

**POST /admin/reset**
- Description: Reset database (only available on dev platform)
- Authentication: None
- Response: 200 OK

## Authentication

### JWT Tokens

Protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

JWT tokens expire after 1 hour. Use the refresh endpoint to obtain a new access token.

### Refresh Tokens

Refresh tokens are valid for 60 days and can be used to obtain new JWT tokens without re-authenticating. Include the refresh token in the Authorization header:
```
Authorization: Bearer <refresh_token>
```

### API Keys

Webhook endpoints require an API key in the Authorization header:
```
Authorization: ApiKey <api_key>
```

## Response Formats

### Chirp Object
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "body": "chirp content",
  "user_id": "uuid"
}
```

### User Object
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

### Login Response
```json
{
  "id": "uuid",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "email": "user@example.com",
  "token": "jwt_token_string",
  "refresh_token": "refresh_token_string",
  "is_chirpy_red": false
}
```

## Features

- User registration and authentication with JWT tokens
- Secure password hashing using Argon2
- Refresh token mechanism for extended sessions
- Create, read, and delete chirps (max 140 characters)
- Filter chirps by author
- Sort chirps in ascending or descending order
- Automatic profanity filtering
- Premium user upgrades via webhooks
- PostgreSQL database with type-safe queries
- Metrics tracking for static file server

## Database Schema

The application uses three main tables:
- `users`: User accounts with email and hashed passwords
- `chirps`: User-generated posts linked to user accounts
- `refresh_tokens`: Active refresh tokens for user sessions

## Project Structure

```
/
├── main.go                # Server setup and routing
├── chirps.go              # Chirp handlers
├── users.go               # User handlers
├── webhooks.go            # Webhook handlers
├── internal/
│   ├── auth/              # Authentication utilities
│   └── database/          # Generated database code
└── sql/
    ├── schema/            # Database migrations
    └── queries/           # SQL queries for code generation
```
