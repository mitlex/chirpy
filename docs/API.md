# Chirpy API Reference

## Table of Contents
- [General](#general)
  - [GET /api/healthz](#get-apihealthz)
- [Authentication & Users](#authentication--users)
  - [POST /api/users](#post-apiusers)
  - [PUT /api/users](#put-apiusers)
  - [POST /api/login](#post-apilogin)
  - [POST /api/refresh](#post-apirefresh)
  - [POST /api/revoke](#post-apirevoke)
- [Chirps](#chirps)
  - [POST /api/chirps](#post-apichirps)
  - [GET /api/chirps](#get-apichirps)
  - [GET /api/chirps/{chirpID}](#get-apichirpschirpid)
  - [DELETE /api/chirps/{chirpID}](#delete-apichirpschirpid)
- [Webhooks](#webhooks)
  - [POST /api/polka/webhooks](#post-apipolkawebhooks)
- [Admin](#admin)
  - [GET /admin/metrics](#get-adminmetrics)
  - [POST /admin/reset](#post-adminreset)

## General

### `GET /api/healthz`

Server readiness and health check. Used to verify the server is running and ready to accept incoming traffic.

#### Authentication
- **Required**: No

#### Request Headers
- None

#### Responses

**`200 OK`**
- **Content-Type**: `text/plain; charset=utf-8`

**Body:**
```text
OK
```

## Authentication & Users

### `POST /api/users`

Creates a new user account with a hashed password.

#### Authentication
- **Required**: No

#### Request Headers
- `Content-Type: application/json`

#### Request Body
| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | Yes | The user's email address. |
| `password` | string | Yes | The user's plaintext password (hashed using Argon2/Bcrypt before database storage). Cannot be empty. |

**Example Request:**
```json
{
  "email": "user@example.com",
  "password": "supersecretpassword"
}
```

#### Responses

**`201 Created`** — User successfully created.
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

**`400 Bad Request`** — Malformed JSON or missing/empty password.
```json
{
  "error": "Password not provided"
}
```

**`500 Internal Server Error`** — Server or database failure (e.g., password hashing failure or database error).
```json
{
  "error": "Error creating user"
}
```

### `PUT /api/users`

Updates the authenticated user's email address and password.

#### Authentication
- **Required**: Yes
- **Header**: `Authorization: Bearer <jwt_token>`

#### Request Headers
- `Content-Type: application/json`

#### Request Body
| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | Yes | The updated email address. |
| `password` | string | Yes | The updated plaintext password (hashed before storing). Cannot be empty. |

**Example Request:**
```json
{
  "email": "updated_user@example.com",
  "password": "newsupersecretpassword"
}
```

#### Responses

**`200 OK`** — User credentials successfully updated.
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:30:00Z",
  "email": "updated_user@example.com",
  "is_chirpy_red": false
}
```

**`400 Bad Request`** — Malformed JSON or missing/empty password.
```json
{
  "error": "Password not provided"
}
```

**`401 Unauthorized`** — Missing, invalid, or expired JWT bearer token.
```json
{
  "error": "invalid token"
}
```

**`500 Internal Server Error`** — Server or database failure.
```json
{
  "error": "Something went wrong"
}
```

### `POST /api/login`

Authenticates a user with an email and password, returning user details alongside access and refresh tokens.

#### Authentication
- **Required**: No

#### Request Headers
- `Content-Type: application/json`

#### Request Body
| Field | Type | Required | Description |
|---|---|---|---|
| `email` | string | Yes | The user's registered email address. |
| `password` | string | Yes | The user's plaintext password. |

**Example Request:**
```json
{
  "email": "user@example.com",
  "password": "supersecretpassword"
}
```

#### Responses

**`200 OK`** — Authentication successful. Returns user profile, access token (JWT valid for 1 hour), and refresh token (valid for 60 days).
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "64_hex_character_refresh_token..."
}
```

**`400 Bad Request`** — Malformed JSON payload.
```json
{
  "error": "Invalid JSON"
}
```

**`401 Unauthorized`** — Invalid credentials (user not found or password mismatch).
```json
{
  "error": "Incorrect email or password"
}
```

**`500 Internal Server Error`** — Server failure generating tokens or accessing the database.
```json
{
  "error": "Server error occurred"
}
```

### `POST /api/refresh`

Issues a new access token (JWT) using a valid, non-expired, and non-revoked refresh token.

#### Authentication
- **Required**: Yes
- **Header**: `Authorization: Bearer <refresh_token>`

#### Request Headers
- None

#### Request Body
- None

#### Responses

**`200 OK`** — Successfully refreshed. Returns a new access token valid for 1 hour.
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**`401 Unauthorized`** — Missing, malformed, expired, or revoked refresh token.
```json
{
  "error": "Invalid refresh token"
}
```

**`500 Internal Server Error`** — Server failure generating token or processing response.
```json
{
  "error": "Server error occurred"
}
```

### `POST /api/revoke`

Revokes a refresh token, preventing it from being used to mint new access tokens.

#### Authentication
- **Required**: Yes
- **Header**: `Authorization: Bearer <refresh_token>`

#### Request Headers
- None

#### Request Body
- None

#### Responses

**`204 No Content`** — Refresh token successfully revoked. No response body.

**`401 Unauthorized`** — Missing or malformed `Authorization` header.
```json
{
  "error": "unauthorized request"
}
```

**`500 Internal Server Error`** — Server or database failure while revoking the token.
```json
{
  "error": "Server error occurred"
}
```

## Chirps

### `POST /api/chirps`

Creates a new chirp associated with the authenticated user.

#### Authentication
- **Required**: Yes
- **Header**: `Authorization: Bearer <jwt_token>`

#### Request Headers
- `Content-Type: application/json`

#### Request Body
| Field | Type | Required | Description |
|---|---|---|---|
| `body` | string | Yes | The content of the chirp. Must not exceed 140 characters. Censored profanity words (`kerfuffle`, `sharbert`, `fornax`) will be replaced with `****`. |

**Example Request:**
```json
{
  "body": "This is a clean chirp, no kerfuffle here!"
}
```

#### Responses

**`201 Created`** — Chirp created successfully.
```json
{
  "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z",
  "body": "This is a clean chirp, no **** here!",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**`400 Bad Request`** — Malformed JSON or chirp exceeds 140 characters.
```json
{
  "error": "Chirp is too long"
}
```

**`401 Unauthorized`** — Missing, invalid, or expired JWT bearer token.
```json
{
  "error": "unauthorized request"
}
```

**`500 Internal Server Error`** — Server failure creating chirp in database or encoding response.
```json
{
  "error": "Chirp creation error"
}
```
*or*
```json
{
  "error": "Server error occurred"
}
```

### `GET /api/chirps`

Retrieves a list of chirps. Supports optional filtering by author and custom sorting order.

#### Authentication
- **Required**: No

#### Request Headers
- None

#### Query Parameters
| Parameter | Type | Required | Description |
|---|---|---|---|
| `author_id` | UUID string | No | Filter chirps created by a specific user UUID. |
| `sort` | string | No | Sort order by `created_at` timestamp. Accepted values: `asc` (default, oldest first) or `desc` (newest first). |

**Example Request URL:**
```http
GET /api/chirps?author_id=123e4567-e89b-12d3-a456-426614174000&sort=desc
```

#### Responses

**`200 OK`** — Returns a list of chirps (returns an empty list `[]` if no chirps match).
```json
[
  {
    "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z",
    "body": "Hello world!",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
]
```

**`400 Bad Request`** — Provided `author_id` query parameter is not a valid UUID format.
```json
{
  "error": "invalid author_id"
}
```

**`500 Internal Server Error`** — Server or database failure.
```json
{
  "error": "Something went wrong"
}
```
*or*
```json
{
  "error": "Server error occurred"
}
```

### `GET /api/chirps/{chirpID}`

Retrieves a single chirp by its UUID.

#### Authentication
- **Required**: No

#### Request Headers
- None

#### Path Parameters
| Parameter | Type | Required | Description |
|---|---|---|---|
| `chirpID` | UUID string | Yes | The unique identifier of the chirp. |

**Example Request URL:**
```http
GET /api/chirps/94b7e44c-3604-42e3-bef7-ebfcc3efff8f
```

#### Responses

**`200 OK`** — Chirp found and returned.
```json
{
  "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
  "created_at": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T12:00:00Z",
  "body": "Hello world!",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

**`400 Bad Request`** — Provided `chirpID` is not a valid UUID format.
```json
{
  "error": "Invalid UUID string"
}
```

**`404 Not Found`** — No chirp exists with the provided ID.
```json
{
  "error": "Chirp not found"
}
```

**`500 Internal Server Error`** — Server or database failure.
```json
{
  "error": "Error getting chirp"
}
```
*or*
```json
{
  "error": "Server error occurred"
}
```

### `DELETE /api/chirps/{chirpID}`

Deletes a chirp by its UUID. Only the author who created the chirp is authorized to delete it.

#### Authentication
- **Required**: Yes
- **Header**: `Authorization: Bearer <jwt_token>`

#### Request Headers
- None

#### Path Parameters
| Parameter | Type | Required | Description |
|---|---|---|---|
| `chirpID` | UUID string | Yes | The unique identifier of the chirp to delete. |

**Example Request URL:**
```http
DELETE /api/chirps/94b7e44c-3604-42e3-bef7-ebfcc3efff8f
```

#### Responses

**`204 No Content`** — Chirp successfully deleted. No response body.

**`400 Bad Request`** — Provided `chirpID` is not a valid UUID format.
```json
{
  "error": "Invalid UUID string"
}
```

**`401 Unauthorized`** — Missing, invalid, or expired JWT bearer token.
```json
{
  "error": "invalid token"
}
```

**`403 Forbidden`** — Authenticated user is not the author of the chirp.
```json
{
  "error": "Unauthorized request"
}
```

**`404 Not Found`** — Chirp with the given ID does not exist.
```json
{
  "error": "Chirp not found"
}
```

**`500 Internal Server Error`** — Server or database failure.
```json
{
  "error": "Error getting chirp"
}
```
*or*
```json
{
  "error": "Server error occurred"
}
```

## Webhooks

### `POST /api/polka/webhooks`

Webhook endpoint for the Polka payment provider. Upgrades a user to Chirpy Red when a `user.upgraded` event is received.

#### Authentication
- **Required**: Yes (API Key)
- **Header**: `Authorization: ApiKey <polka_api_key>`

#### Request Headers
- `Content-Type: application/json`

#### Request Body
| Field | Type | Required | Description |
|---|---|---|---|
| `event` | string | Yes | Event name. Only `user.upgraded` triggers an upgrade; all other events return `204 No Content` without taking action. |
| `data.user_id` | UUID string | Yes | The UUID of the user to upgrade. |

**Example Request:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

#### Responses

**`204 No Content`** — Webhook received and processed successfully (or event ignored). No response body.

**`400 Bad Request`** — Malformed JSON payload.
```json
{
  "error": "Invalid JSON"
}
```

**`401 Unauthorized`** — Missing, malformed, or incorrect Polka API key in the `Authorization` header.
```json
{
  "error": "unauthorized request"
}
```

**`404 Not Found`** — User specified in `data.user_id` was not found.
```json
{
  "error": "User not found"
}
```

**`500 Internal Server Error`** — Server or database failure.
```json
{
  "error": "Something went wrong"
}
```

## Admin

### `GET /admin/metrics`

Renders an HTML dashboard displaying the total number of visits to the server assets/fileserver.

#### Authentication
- **Required**: No

#### Request Headers
- None

#### Responses

**`200 OK`**
- **Content-Type**: `text/html; charset=utf-8`

**Example Response Body:**
```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 42 times!</p>
  </body>
</html>
```

### `POST /admin/reset`

Resets the server hits counter back to 0 and deletes all users (along with cascade-associated chirps and refresh tokens) from the database. Only permitted when running in a `dev` environment.

#### Authentication
- **Required**: No

#### Request Headers
- None

#### Request Body
- None

#### Responses

**`200 OK`** — Metrics counter reset and database purged successfully.
- **Content-Type**: `text/plain; charset=utf-8`
- No response body.

**`403 Forbidden`** — Server is not running with `PLATFORM=dev`.
```json
{
  "error": "forbidden request"
}
```

**`500 Internal Server Error`** — Server failed to reset user records in the database.
```json
{
  "error": "Error resetting users"
}
```