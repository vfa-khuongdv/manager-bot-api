# Golang CMS

This is a Golang project designed to handle a simple web service with user management, roles, permissions, and refresh tokens. It uses MySQL for the database, Docker for containerization, and includes support for migrations, seeding, and authentication.

## Project Structure

The project follows a clean architecture and is organized into the following directories:

```
├── Dockerfile                        # Docker configuration for the application
├── README.md                         # Project documentation
├── cmd                               # Command-line interfaces (CLI)
│   ├── seeder                        # Seeder for initial data population
│   │   └── seeder.go
│   └── server                        # Main entry point for the web server
│       └── main.go
├── docker-compose.yml                # Docker Compose configuration for the app and MySQL
├── docs                              # API documentation
│   └── api_spec.md
├── go.mod                            # Go module dependencies
├── go.sum                            # Go module checksums
├── internal                          # Core application logic
│   ├── configs                       # Configuration files for database, environment variables, JWT, etc.
│   ├── constants                     # Constants and error handling
│   ├── database                      # Database migrations and seeding
│   ├── handlers                      # HTTP request handlers
│   ├── middlewares                   # Middlewares for authentication and logging
│   ├── models                        # Data models for the application
│   ├── repositories                  # Repositories for database access
│   ├── routes                        # Routes and routing logic
│   ├── services                      # Business logic for authentication, user, etc.
│   └── utils                         # Utility functions (e.g., for encryption, validation)
├── pkg                               # External packages
│   ├── logger                        # Logger utility
│   └── mailer                        # Mailer for sending emails
├── tests                             # Unit and integration tests
│   └── internal/utils
│       └── security_test.go
```

## Prerequisites

Before getting started, ensure that you have the following installed:

- [Go](https://golang.org/dl/) (Go 1.23 or later)
- [Docker](https://www.docker.com/products/docker-desktop)
- [Docker Compose](https://docs.docker.com/compose/)
- [MySQL](https://www.mysql.com/)
- [Migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for database migrations)

## Setup Instructions

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/yourproject.git
cd yourproject
cp .env.example .env
```

### 2. Build and run the application using Docker

You can use Docker Compose to set up both the app and the MySQL database:

```bash
docker-compose up --build
```

This will:

- Build the Docker images.
- Start a MySQL container.
- Start the application container.

### 3. Database Migrations

To create a new migration file, use the following command:

```bash
migrate create -ext sql -dir internal/database/migrations -seq your_migration_name
```

For example, to create a feedback table migration:
```bash
migrate create -ext sql -dir internal/database/migrations -seq feedback_table
```

This will create two files:
- XXXXXX_feedback_table.up.sql (for applying the migration)
- XXXXXX_feedback_table.down.sql (for reverting the migration)

The project includes migrations for creating the necessary tables in the MySQL database.
To apply the migrations:

```bash
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db_2" up
```

To revert migrations, you can use the down command:

```bash
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db_2" down
```

You can also revert a specific number of migrations by adding the number after the down command:

```bash
migrate -path ./internal/database/migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/golang_db_2" down 1
```

This will run the migration scripts and populate the database.

### 4. Seeding the Database

To seed the database with initial data (e.g., default users, roles, permissions), run:

```bash
docker-compose exec app go run cmd/seeder/seeder.go
```

### 5. Running the Server

There are two ways to run the server:

#### Using Air (Recommended for Development)

[Air](https://github.com/air-verse/air) provides live-reloading capability which is great for development. To use it:

1. Install Air:
```bash
# binary will be $(go env GOPATH)/bin/air
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# or install it into ./bin/
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s

air -v
```

2. If the `air` command is not found after installation, add this to your `~/.bash_profile` or `~/.zshrc`:
```bash
export PATH=$PATH:$HOME/go/bin
```

3. Run the server with live-reloading:
```bash
air
```

#### Direct Go Run (Alternative)

If you prefer to run the server directly without live-reloading:

```bash
go run cmd/server/main.go
```

The server will start and be available at `http://localhost:3000`.

### 6. phpmyadmin

PHPMyAdmin is available for database management through a web interface at:
- URL: `http://localhost:8080`
- Username: `root`
- Password: `root`

## Environment Variables

The following environment variables are required for the application:

Database Configuration:
- `DB_HOST` - MySQL database host
- `DB_USERNAME` - MySQL database username
- `DB_PASSWORD` - MySQL database password
- `DB_DATABASE` - MySQL database name
- `DB_PORT` - MySQL port number

JWT Configuration:
- `JWT_SECRET_KEY` - Secret key for JWT token generation
- `JWT_EXPIRES_IN` - JWT token expiration time (e.g., "24h")

Server Configuration:
- `SERVER_PORT` - Port number for the application server (default: 3000)
- `SERVER_MODE` - Server mode ("development" or "production")

Mail Configuration (if using email features):
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port
- `SMTP_USERNAME` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `SMTP_FROM_ADDRESS` - Email address used as sender

Redis Configuration (for caching and session management):
- `REDIS_HOST` - Redis server host
- `REDIS_PORT` - Redis server port
- `REDIS_PASSWORD` - Redis password (if any)

These can be set in the `.env` file or passed directly as environment variables. A sample `.env.example` file is provided in the repository.

Check the `docs/api_spec.md` for a detailed API specification.

## Testing

Run unit tests with the following command:

```bash
go test ./...
```

For specific tests, use:

```bash
go test -v path/to/test
```

### Unit Tests Directory

The test files are located under the `tests` directory. The tests follow the Go testing conventions.

## Contribution Guidelines

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/feature-name`).
3. Commit your changes (`git commit -am 'Add feature'`).
4. Push to the branch (`git push origin feature/feature-name`).
5. Open a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Features

### User Management
- User registration and authentication
- Role-based access control
- Permission management

### Settings Management
- Application settings configuration
- System preferences

### Reminder Schedules
The system includes a reminder scheduling feature that allows you to set up automated messages to be sent to Chatwork rooms based on cron expressions.

#### Reminder Schedule Features
- Create, read, update, and delete reminder schedules
- Schedule messages using cron expressions
- Enable/disable schedules without deleting them
- Associate schedules with specific projects
- Send messages to Chatwork rooms

#### Setting Up Reminders
1. Create a project or use an existing one
2. Create a reminder schedule with:
   - A valid cron expression (e.g., `0 9 * * 1-5` for 9 AM every weekday)
   - Chatwork room ID
   - Chatwork API token
   - Message content
3. Enable the schedule

#### Cron Expression Format
The system uses standard cron expression format with five fields:
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of the month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

Examples:
- `0 9 * * 1-5`: At 9:00 AM, Monday through Friday
- `0 17 * * 5`: At 5:00 PM on Friday
- `0 14 25 * *`: At 2:00 PM on the 25th of each month

### Key Sections:
1. **Project Structure**: A breakdown of the directories and files with brief descriptions.
2. **Setup Instructions**: Instructions for setting up the project locally, including dependencies and Docker setup.
3. **Environment Variables**: Key environment variables needed for the project to run properly.
4. **Features**: Core features of the application including user management, settings, and reminder schedules.
5. **Testing**: How to run unit tests in the project.
6. **Contribution Guidelines**: Instructions for contributing to the project.
