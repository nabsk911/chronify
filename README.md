# Chronify Backend

Chronify is a robust backend service built in Go, designed to manage timelines, events, users, and bookmarks. It features an AI-powered event generation system using Google's Gemini AI, allowing users to effortlessly create and update complex timelines.

## Features

- **User Authentication**: Secure registration and login using bcrypt for password hashing and JWT for session management.
- **Timelines Management**: Create, read, update, and delete custom timelines.
- **Timeline Discovery**: Search functionality and support for public/private timelines.
- **Event Management**: Add, organize, and delete specific events within a timeline. Supports bulk operations.
- **AI-Powered Events**: Seamlessly generate or modify events in a timeline using prompts powered by the Gemini AI API.
- **Bookmarking**: Bookmark favorite timelines for quick access.
- **Rate Limiting**: Custom middleware to prevent abuse, specifically on AI generation endpoints.
- **Database Driven**: PostgreSQL backing with auto-generated typesafe queries via `sqlc`.

## Tech Stack

- **Language**: Go 1.25.0
- **Database**: PostgreSQL (using `pgx` driver)
- **Database Queries**: [sqlc](https://sqlc.dev/) for type-safe database interactions
- **AI Integration**: Google GenAI SDK (`google.golang.org/genai`)
- **Authentication**: JWT (`github.com/golang-jwt/jwt/v5`)
- **Environment Management**: Godotenv (`github.com/joho/godotenv`)

## Project Structure

The project follows a standard Go application layout for a clean separation of concerns:

```text
├── internal/
│   ├── app/          # Core application configuration and initialization
│   ├── auth/         # Authentication logic (JWT generation, password hashing)
│   ├── db/           # Auto-generated sqlc database models and queries
│   ├── handlers/     # HTTP route handlers (Users, Timelines, Events, Bookmarks)
│   ├── middleware/   # HTTP middlewares (Auth validation, CORS, Rate limiting)
│   ├── routes/       # API route definitions and multiplexing
│   └── utils/        # Shared utilities (JSON formatting, validators, mappers)
├── sql/
│   ├── queries/      # Raw SQL queries used by sqlc to generate Go code
│   └── schema/       # Database migrations (Goose format)
├── .env.example      # Example environment variables
├── go.mod            # Go module definitions
├── main.go           # Application entrypoint
└── sqlc.yaml         # Configuration for sqlc code generation
```

## API Endpoints

### Authentication
- `POST /register` - Register a new user
- `POST /login` - Login and receive a JWT token

### Timelines
- `GET /timelines` - Get all timelines for the authenticated user
- `POST /timelines` - Create a new timeline
- `GET /timelines/{timelineId}` - Get a specific timeline
- `PUT /timelines/{timelineId}` - Update a timeline
- `DELETE /timelines/{timelineId}` - Delete a timeline
- `GET /timelines/search` - Search through user's timelines
- `GET /timelines/public` - Get public timelines
- `GET /timelines/public/search` - Search through public timelines

### Events
- `GET /timelines/{timelineId}/events` - Get all events for a timeline
- `POST /timelines/{timelineId}/events` - Upsert events into a timeline
- `DELETE /timelines/{timelineId}/events/{eventId}` - Delete a specific event

### AI Events
- `POST /timelines/{timelineId}/aievents` - Generate new timeline events via AI prompt
- `PUT /timelines/{timelineId}/aievents` - Update existing timeline events via AI prompt

### Bookmarks
- `GET /bookmarks` - Get all bookmarked timelines
- `POST /bookmarks/{timelineId}` - Add a bookmark
- `DELETE /bookmarks/{timelineId}` - Remove a bookmark

## Getting Started

### Prerequisites
- Go 1.25.0+
- PostgreSQL database
- Gemini API Key

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up the environment variables. Create a `.env` file in the root directory:
   ```env
   DATABASE_URL=postgres://user:pass@localhost:5432/chronify
   JWT_SECRET=your_jwt_secret
   GEMINI_API_KEY=your_gemini_api_key
   PORT=8080
   ```
4. Run database migrations (using [goose](https://github.com/pressly/goose)):
   ```bash
   goose -dir sql/schema postgres "postgres://user:pass@localhost:5432/chronify" up
   ```
5. (Optional) Generate sqlc models if making changes to `sql/queries/`:
   ```bash
   sqlc generate
   ```
6. Start the server:
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`.
