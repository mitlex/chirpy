# Chirpy

Chirpy is a RESTful HTTP server and mock social network built from scratch in Go. It handles client requests with JSON payloads, standard HTTP status codes, and custom headers.

> **Note:** This project was built as part of Boot.dev's Back-End Development Path, following guided instruction.

## Learning Outcomes

- **Production HTTP & Server Architecture**: Built a Go REST server from scratch using standard library routing (`net/http`), custom middleware for request/metrics tracking, and HTTP file servers for static asset delivery.
- **RESTful API Design & Contracts**: Designed standardized REST contracts communicating via JSON payloads, custom headers, query parameters (sorting and filtering), status codes, and structured error handling.
- **State & Environment Management**: Managed centralized application state across handlers, controlling database query access, environment permissions, and runtime secrets (`.env`).
- **Authentication & Cryptography**: Implemented secure JWT access tokens, database-backed refresh tokens, and password hashing using Argon2id.
- **Database Engineering**: Designed relational schemas and managed migrations using Goose, executing type-safe SQL queries against PostgreSQL.
- **Webhooks & Third-Party APIs**: Handled external payment provider webhooks authenticated via custom API keys.
- **API Documentation & Tooling**: Documented comprehensive API reference contracts in Markdown and validated HTTP payloads using the VS Code REST Client extension and unit tests (`go test ./...`).

## Getting Started

### Prerequisites
- Go 1.26.5 (may work with earlier versions, untested)
- PostgreSQL 15+ (see installation instructions below) 
    - Note: version 16+ was used for this project, but 15+ should work just fine.

### Installation/Execution

```sh
git clone https://github.com/mitlex/chirpy
cd chirpy
go install
```

OR simply download and install the module directly with this one-liner:
```sh
go install github.com/mitlex/chirpy@latest
```

#### PostgreSQL Installation/Setup

**NOTE**: any macOS instructions are untested.

##### 1. Install PostgreSQL

**macOS** (with Homebrew):
```sh
brew install postgresql@15
```

**Linux / WSL (Debian-based):**
```sh
sudo apt update
sudo apt install postgresql postgresql-contrib
```

##### 2. Verify Installation

```sh
psql --version
```

Confirm the output shows version 15 or higher.

##### 3. (Linux/WSL only) Set the postgres system user password

**This is your postgres system level password so don't forget it!**

```sh
sudo passwd postgres
```

##### 4. Start the PostgreSQL server

- **macOS:** `brew services start postgresql@15`
- **Linux:** `sudo service postgresql start`

##### 5. Connect to PostgreSQL

- **macOS:** `psql postgres`
- **Linux:** `sudo -u postgres psql`

You should land in a prompt like:
```
postgres=#
```

##### 6. Create the database

```sql
CREATE DATABASE chirpy;
```

##### 7. Connect to the new database

```sql
\c chirpy
```

Prompt should update to:
```
chirpy=#
```

##### 8. (Linux/WSL only) Set the database user password

This is the **database user's** password; replace password_here with something easy to remember.

```sql
ALTER USER postgres WITH PASSWORD 'password_here';
```

##### 9. Test the connection to the database

```sql
SELECT version();
```

Type \`exit\` to leave the \`psql\` shell when done.

##### 10. Database Migrations

Install [Goose](https://github.com/pressly/goose) (if not already installed):

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Run migrations to apply the database schema:

```sh
cd sql/schema
goose postgres "postgres://postgres:password_here@localhost:5432/chirpy" up
cd ../..
```

### Configuration

Create a `.env` file in the root directory:

```env
DB_URL=postgres://postgres:password_here@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
JWT_SECRET=your_super_secret_jwt_key
POLKA_KEY=your_polka_api_key
```

### Running the Program

```sh
go build -o chirpy
./chirpy
```

OR build and run in one step:
```sh
go run .
```

I recommend running the program in one terminal and sending requests to the server in a separate terminal, or by using a 3rd party HTTP request tool of your choosing (e.g. curl).

## Testing

Testing was conducted via unit tests (`go test ./...`) for cryptographic and authentication packages, alongside end-to-end HTTP request validation against a local PostgreSQL database instance.

## API Reference

A brief overview of available endpoints. For full payload schemas, headers, and error codes, see [API.md](./docs/API.md).

| Method | Endpoint | Auth Required | Description |
|---|---|---|---|
| **General** | | | |
| `GET` | `/api/healthz` | No | Server readiness / health check |
| **Authentication & Users** | | | |
| `POST` | `/api/users` | No | Create a new user |
| `PUT` | `/api/users` | Yes (JWT) | Update user email and password |
| `POST` | `/api/login` | No | Authenticate user and receive access/refresh tokens |
| `POST` | `/api/refresh` | Yes (Refresh Token) | Refresh an expired access token |
| `POST` | `/api/revoke` | Yes (Refresh Token) | Revoke a refresh token |
| **Chirps** | | | |
| `POST` | `/api/chirps` | Yes (JWT) | Create a new chirp |
| `GET` | `/api/chirps` | No | Retrieve chirps (supports sorting and author filter) |
| `GET` | `/api/chirps/{chirpID}` | No | Retrieve a specific chirp by ID |
| `DELETE` | `/api/chirps/{chirpID}` | Yes (JWT) | Delete a chirp (author only) |
| **Webhooks** | | | |
| `POST` | `/api/polka/webhooks` | Yes (API Key) | Handle Polka webhook to upgrade user to Chirpy Red |
| **Admin** | | | |
| `GET` | `/admin/metrics` | No | View server hits and metrics |
| `POST` | `/admin/reset` | No | Reset hit counter and purge database records |

## Future Improvements

### Features
- **Pagination**: Limit returned chirps using cursor-based or offset pagination.
- **Social Graph**: Follow/block systems and a personalized "following" feed.
- **Interactions**: Like counters and user profiles (bios, avatars).
- **Search**: Full-text search for chirps and hashtags.

### Engineering & Infrastructure
- **Containerization**: Dockerize the application and database with `docker-compose`.
- **Caching**: Implement a Redis caching layer for read-heavy endpoints.
- **Testing**: Expand integration test coverage using Go's `httptest` package.
- **CI/CD**: Set up GitHub Actions for automated linting (`golangci-lint`) and test execution.

## Acknowledgments

This project was developed as part of the [Back-End Development Path](https://www.boot.dev/tracks/backend) on [Boot.dev](https://www.boot.dev). 

Thanks to the authors at Boot.dev!