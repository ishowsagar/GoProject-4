# 🛠️ Setup and Operations Guide

Complete guide for setting up and running this project from scratch, including what to do after restarting your PC.

---

## 📦 Initial Setup (First Time Only)

### 1. Install Required Tools

#### Go Installation

```bash
# Download from https://golang.org/dl/
# Verify installation
go version
```

#### Docker Desktop Installation

```bash
# Download from https://www.docker.com/products/docker-desktop
# Start Docker Desktop application
# Verify installation
docker --version
docker compose version
```

#### Goose (Database Migration Tool)

```bash
# Install goose globally
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify installation
goose --version
```

#### Air (Hot Reload Tool)

```bash
# Install air globally
go install github.com/air-verse/air@latest

# Verify installation
air -v
```

### 2. Clone/Create Project

```bash
# Navigate to your projects directory
cd C:\Users\asus\Documents\LearnCode

# If cloning from Git
git clone <your-repo-url> GoProject#4

# Navigate into project
cd GoProject#4
```

### 3. Install Go Dependencies

```bash
# Download all dependencies from go.mod
go mod download

# Verify dependencies
go mod verify
```

---

## 🐘 PostgreSQL Setup with Docker

### Understanding docker-compose.yml

The project uses **two PostgreSQL databases**:

1. **Main Database** (port 5445) - For development/production
2. **Test Database** (port 5500) - For running tests

**Why custom ports?**

- Windows reserves ports 5345-5444 for dynamic allocation
- Standard PostgreSQL port 5432 conflicts with Windows
- Using 5445 and 5500 avoids conflicts

### docker-compose.yml Configuration

```yaml
services:
  db:
    image: postgres:12.4-alpine
    ports:
      - "5445:5432" # External:Internal
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_USER: postgres
      POSTGRES_DB: postgres
    volumes:
      - ./database/postgres-data:/var/lib/postgresql/data
    restart: unless-stopped

  test_db:
    image: postgres:12.4-alpine
    ports:
      - "5500:5432"
    environment:
      POSTGRES_PASSWORD: postgres
      POSTGRES_USER: postgres
      POSTGRES_DB: postgres
    volumes:
      - ./database/postgres-test-data:/var/lib/postgresql/data
    restart: unless-stopped
```

### First Time Database Setup

```bash
# Make sure Docker Desktop is running first!

# Start PostgreSQL containers in background (-d = detached)
docker compose up -d

# Verify containers are running
docker ps

# You should see:
# - goproject4-db-1 (port 5445)
# - goproject4-test_db-1 (port 5500)
```

### Database Connections Strings

**Main Database:**

```
postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable
```

**Test Database:**

```
postgres://postgres:postgres@localhost:5500/postgres?sslmode=disable
```

**Format Breakdown:**

```
postgres://[user]:[password]@[host]:[port]/[database]?sslmode=disable
```

---

## 🔄 Database Migrations (Goose)

### What Are Migrations?

Migrations are versioned SQL files that modify your database schema over time:

- Each file numbered (00001, 00002, etc.)
- Contains "Up" (apply changes) and "Down" (rollback changes)
- Tracked by goose in `goose_db_version` table

### Migration Files in This Project

```
migrations/
  ├── 00001_users.sql           # Creates users table
  ├── 00002_workouts.sql        # Creates workouts table
  ├── 00003_workout_entries.sql # Creates workout_entries table
  ├── 00004_tokens.sql          # Creates tokens table
  └── 00005_user_id_alter.sql   # Adds user_id to workouts
```

### Running Migrations

#### Apply All Pending Migrations

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" up
```

#### Check Migration Status

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" status
```

#### Rollback Last Migration

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" down
```

#### Reset Database (Rollback All)

```bash
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" reset
```

#### Fix Migration Issues

```bash
# If migrations are out of sync
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" fix
```

#### Create New Migration

```bash
goose -dir migrations create add_new_feature sql
```

### Migration File Structure

```sql
-- +goose Up
-- Apply changes here
CREATE TABLE example (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- +goose Down
-- Rollback changes here
DROP TABLE IF EXISTS example;
```

### Automatic Migration on App Start

This project runs migrations automatically when the app starts (see `internal/store/database.go`):

```go
func (s *Storage) Migratefs() error {
    goose.SetBaseFS(migrations.FS)
    if err := goose.Up(s.db, "."); err != nil {
        return err
    }
    return nil
}
```

**Benefit:** No need to manually run migrations before starting the app!

---

## 🚀 Running the Application

### Development Mode (with Air)

Air watches for file changes and automatically rebuilds/restarts the server.

```bash
# Start in hot reload mode
air

# Air will:
# 1. Build the application
# 2. Start the server on port 8080
# 3. Watch for .go file changes
# 4. Rebuild automatically when files change
```

### Production Mode (Manual Build)

```bash
# Build the binary
go build -o workout-api .

# Run the binary
./workout-api

# Run with custom port
./workout-api -port=8080
```

### Verify Server is Running

```bash
# Check server health
curl http://localhost:8080/health

# Expected response:
# {"data":{"health":"up to the moon"}}
```

---

## 🔄 After PC Restart / Shutdown

When you restart your PC, Docker containers stop. Here's how to get everything running again:

### Complete Restart Procedure

```bash
# Step 1: Open Docker Desktop
# Make sure Docker Desktop is running (check system tray)

# Step 2: Navigate to project directory
cd C:\Users\asus\Documents\LearnCode\GoProject#4

# Step 3: Start Docker containers
docker compose up -d

# Step 4: Verify containers are running
docker ps

# You should see both containers running:
# - goproject4-db-1 (0.0.0.0:5445->5432/tcp)
# - goproject4-test_db-1 (0.0.0.0:5500->5432/tcp)

# Step 5: Start the application
air

# That's it! Server will be running on http://localhost:8080
```

### Quick Status Checks

```bash
# Check if Docker is running
docker ps

# Check if port 8080 is in use
netstat -ano | findstr :8080

# Check if databases are accessible
docker exec -it goproject4-db-1 psql -U postgres -c "SELECT version();"
```

---

## 🐛 Common Issues and Solutions

### Issue 1: Docker Containers Won't Start

**Symptom:**

```
Error response from daemon: Ports are not available
```

**Solution:**

```bash
# Check what's using the ports
netstat -ano | findstr "5445 5500"

# Kill process if needed (replace PID)
taskkill /PID <process_id> /F

# Or use PowerShell
$process = Get-NetTCPConnection -LocalPort 5445 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess
if ($process) { Stop-Process -Id $process -Force }
```

### Issue 2: Port 8080 Already in Use

**Symptom:**

```
listen tcp :8080: bind: Only one usage of each socket address
```

**Solution:**

```bash
# Find process using port 8080
netstat -ano | findstr :8080

# Kill the process (usually a stale air/main process)
taskkill /PID <process_id> /F

# Or use PowerShell
$process = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess
if ($process) { Stop-Process -Id $process -Force }

# Then restart
air
```

### Issue 3: Can't Connect to Database

**Symptom:**

```
dial tcp [::1]:5445: connectex: No connection could be made
```

**Solutions:**

```bash
# 1. Check if Docker containers are running
docker ps

# 2. If not running, start them
docker compose up -d

# 3. Check container logs for errors
docker logs goproject4-db-1

# 4. Restart containers if needed
docker compose restart

# 5. If still not working, recreate containers
docker compose down
docker compose up -d
```

### Issue 4: Migration Errors

**Symptom:**

```
goose: no such table: goose_db_version
```

**Solution:**

```bash
# This is normal for first run. Just run migrations:
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" up
```

**Symptom:**

```
ERROR: column "user_id" already exists
```

**Solution:**

```bash
# Fix migration state
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" fix

# Or check migration status
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" status
```

### Issue 5: Air Not Found

**Symptom:**

```
air: command not found
```

**Solution:**

```bash
# Reinstall air
go install github.com/air-verse/air@latest

# Make sure Go bin is in PATH
# Add to PATH: C:\Users\<YourUsername>\go\bin

# Or run directly
go run github.com/air-verse/air@latest
```

### Issue 6: Go Dependencies Missing

**Symptom:**

```
package not found
```

**Solution:**

```bash
# Download all dependencies
go mod download

# Clean and reinstall
go clean -modcache
go mod download

# Tidy up (removes unused, adds missing)
go mod tidy
```

---

## 🧹 Cleanup Commands

### Stop Everything

```bash
# Stop the Go server
# Press Ctrl+C in the air terminal

# Stop Docker containers
docker compose stop

# Stop and remove containers (keeps data)
docker compose down
```

### Clean Database (Start Fresh)

```bash
# Stop containers and remove volumes (deletes all data!)
docker compose down -v

# Remove data directories
rm -rf database/postgres-data
rm -rf database/postgres-test-data

# Start fresh
docker compose up -d

# Run migrations
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" up
```

### Remove Docker Images

```bash
# List images
docker images

# Remove postgres image
docker rmi postgres:12.4-alpine

# Clean up unused images
docker image prune -a
```

---

## 📋 Daily Development Workflow

### Starting Work

```bash
# 1. Ensure Docker Desktop is running

# 2. Navigate to project
cd C:\Users\asus\Documents\LearnCode\GoProject#4

# 3. Check if containers are running
docker ps

# 4. If not running, start them
docker compose up -d

# 5. Start development server
air

# 6. Start coding!
```

### During Development

```bash
# Air automatically rebuilds when you save .go files

# If you add new dependencies:
go get github.com/some/package
go mod tidy

# If you need to create a migration:
goose -dir migrations create feature_name sql
# Edit the new SQL file
# Restart air to apply migration
```

### Ending Work

```bash
# 1. Stop air (Ctrl+C)

# 2. Optional: Stop Docker containers
docker compose stop

# Database data is persisted, so you can safely shutdown
```

---

## 🔍 Useful Inspection Commands

### Docker Commands

```bash
# See all containers (running and stopped)
docker ps -a

# View container logs
docker logs goproject4-db-1
docker logs goproject4-test_db-1

# Follow logs in real-time
docker logs -f goproject4-db-1

# Execute command in container
docker exec -it goproject4-db-1 psql -U postgres

# Inspect container details
docker inspect goproject4-db-1

# View container resource usage
docker stats
```

### Database Commands (Inside Container)

```bash
# Connect to database
docker exec -it goproject4-db-1 psql -U postgres

# Inside psql:
\l                      # List databases
\dt                     # List tables
\d users                # Describe users table
\d+ workouts            # Detailed table info
SELECT * FROM users;    # Query data
\q                      # Exit psql
```

### Check Ports

```bash
# Windows CMD
netstat -ano | findstr "8080 5445 5500"

# PowerShell
Get-NetTCPConnection -LocalPort 8080,5445,5500 -ErrorAction SilentlyContinue
```

### Check Processes

```bash
# PowerShell
Get-Process -Name "air","main","docker" -ErrorAction SilentlyContinue

# Kill stuck processes
Get-Process -Name "air","main" -ErrorAction SilentlyContinue | Stop-Process -Force
```

---

## 📝 Environment Configuration

### Database Configuration (in code)

Located in `internal/store/database.go`:

```go
func Open() (*Storage, error) {
    db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable")
    // ...
}
```

### Server Configuration

Default port is 8080, can be changed via flag:

```bash
# Run on different port
air -- -port=3000

# Or for built binary
./workout-api -port=3000
```

### Test Database Configuration

Located in `internal/store/workout_store_test.go`:

```go
connectionString := "postgres://postgres:postgres@localhost:5500/postgres?sslmode=disable"
```

---

## 🚦 Startup Checklist

Use this checklist after every restart:

- [ ] Docker Desktop is running
- [ ] Navigate to project directory
- [ ] Run `docker compose up -d`
- [ ] Verify with `docker ps` (2 containers running)
- [ ] Run `air`
- [ ] Test with `curl http://localhost:8080/health`
- [ ] Everything is ready! 🎉

---

## 🔗 Connection Summary

| Service       | Port | Connection String                                                      |
| ------------- | ---- | ---------------------------------------------------------------------- |
| Main Database | 5445 | `postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable` |
| Test Database | 5500 | `postgres://postgres:postgres@localhost:5500/postgres?sslmode=disable` |
| Go Server     | 8080 | `http://localhost:8080`                                                |

---

## 📚 Tool Documentation Links

- **Go**: https://golang.org/doc/
- **Docker**: https://docs.docker.com/
- **Docker Compose**: https://docs.docker.com/compose/
- **PostgreSQL**: https://www.postgresql.org/docs/
- **Goose**: https://github.com/pressly/goose
- **Air**: https://github.com/air-verse/air
- **Chi Router**: https://github.com/go-chi/chi

---

## 💡 Pro Tips

1. **Always start Docker Desktop first** - Nothing will work without it!

2. **Use `docker compose up -d`** - The `-d` flag runs in background so you can use the terminal

3. **Check `docker ps` regularly** - Make sure containers are actually running

4. **Use Air for development** - It's much faster than manual rebuilds

5. **Migrations run automatically** - No need to manually run goose before starting the app

6. **Data persists across restarts** - Your database data is safe in the `database/` folder

7. **Test database is separate** - Running tests won't affect your development data

8. **Keep Docker Desktop running** - You can minimize it, but don't close it

9. **Check logs if something breaks** - `docker logs goproject4-db-1` shows database errors

10. **Use the health endpoint** - Quick way to verify the server is running: `curl http://localhost:8080/health`

---

## 🎯 Quick Command Reference

```bash
# Start everything
docker compose up -d && air

# Stop everything
docker compose stop

# Restart database
docker compose restart db

# View logs
docker logs -f goproject4-db-1

# Run migrations manually
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" up

# Check migration status
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5445/postgres?sslmode=disable" status

# Connect to database
docker exec -it goproject4-db-1 psql -U postgres

# Kill stuck processes
Get-Process -Name "air","main" -ErrorAction SilentlyContinue | Stop-Process -Force

# Check if ports are free
netstat -ano | findstr "8080 5445 5500"

# Health check
curl http://localhost:8080/health
```

---

**Remember:** After each PC restart, just run `docker compose up -d` then `air`. That's it! 🚀
