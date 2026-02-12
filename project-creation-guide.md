# 🏗️ Building a Workout Tracker API from Scratch

## Overview

This guide explains **HOW** to approach building this REST API project step-by-step, focusing on the thought process and architecture decisions rather than specific syntax.

---

## 📋 Phase 1: Project Planning & Setup

### Step 1.1: Define Requirements

**What are we building?**

- Workout tracking system where users can:
  - Create accounts (authentication)
  - Log workouts with multiple exercises
  - Track reps, sets, duration, weight
  - Update and delete their own workouts

**Technical Requirements:**

- RESTful API (JSON responses)
- User authentication (token-based)
- Database persistence
- Authorization (users own their data)

### Step 1.2: Design Database Schema

**Think about relationships:**

- Users have many workouts (1:N)
- Workouts have many entries (1:N)
- Tokens belong to users (N:1)

**Sketch out tables:**

```
users: id, username, email, password_hash, bio, timestamps
workouts: id, user_id, title, description, duration, calories
workout_entries: id, workout_id, exercise_name, sets, reps, weight, notes
tokens: hash, user_id, expiry, scope
```

### Step 1.3: Choose Tech Stack

**Backend Framework:** Go with chi router

- Why? Fast, lightweight, good for learning HTTP fundamentals

**Database:** PostgreSQL

- Why? Reliable, supports transactions, foreign keys

**Migration Tool:** Goose

- Why? Version control for database changes

**Development Tools:**

- Air for hot reload during development
- Docker for running PostgreSQL locally

---

## 📦 Phase 2: Project Structure Setup

### Step 2.1: Initialize Go Module

Create project directory and initialize:

- Set up go.mod file
- Choose module name (github.com/yourname/workout-api)

### Step 2.2: Design Folder Structure

**Think in layers:**

```
/cmd or root
  └─ main.go (entry point)

/internal (private application code)
  ├─ app/ (application bootstrap)
  ├─ api/ (HTTP handlers)
  ├─ store/ (database operations)
  ├─ middleware/ (authentication, logging)
  ├─ routes/ (route configuration)
  ├─ tokens/ (token generation logic)
  └─ utils/ (helpers like JSON writer)

/migrations (SQL files for database schema)
  └─ 00001_users.sql, 00002_workouts.sql, etc.
```

**Why this structure?**

- Separation of concerns (each folder has one job)
- Easy to test (can mock store layer)
- Scalable (add features without breaking existing code)

### Step 2.3: Set Up Docker for PostgreSQL

Create docker-compose.yml:

- Define PostgreSQL service
- Set up volumes for data persistence
- Map ports (5445 to avoid Windows conflicts)
- Add test database on different port

---

## 🎯 Phase 3: Core Application Setup

### Step 3.1: Create Main Entry Point

**main.go responsibilities:**

1. Parse command line flags (port number)
2. Initialize application
3. Set up HTTP server with timeouts
4. Handle graceful shutdown
5. Listen and serve

**Key decision:** Keep main.go thin, move logic to app package

### Step 3.2: Build Application Bootstrap

**internal/app/app.go:**

Create Application struct to hold all dependencies:

- Logger (for error tracking)
- Database connection
- All handlers (workout, user, token)
- Middleware

**NewApplication function flow:**

1. Open database connection
2. Run migrations automatically
3. Initialize stores (workout, user, token)
4. Initialize handlers with their dependencies
5. Set up middleware
6. Return wired-up Application struct

**Why this approach?**

- Dependency injection (everything is explicit)
- Easy to test (can inject mocks)
- Single source of truth for app setup

---

## 🗄️ Phase 4: Database Layer (Store)

### Step 4.1: Design Store Interface Pattern

**Problem:** How to keep code flexible?

**Solution:** Use interfaces!

Define what operations you need:

- WorkoutStore interface (CreateWorkout, GetWorkoutByID, etc.)
- UserStore interface (CreateUser, GetUserByUsername, etc.)
- TokenStore interface (Insert, CreateNewToken, etc.)

Then create PostgreSQL implementations.

**Benefit:** Can swap PostgreSQL for MySQL without changing handlers!

### Step 4.2: Implement Database Connection

**internal/store/database.go:**

- Open function: Create connection pool
- Connection string with credentials
- Use pgx driver (fast PostgreSQL driver)

### Step 4.3: Implement Migration System

**Why migrations?**

- Track database changes over time
- Apply schema updates automatically
- Can rollback if needed

**How it works:**

- Numbered SQL files (00001, 00002, etc.)
- Each has "Up" (apply) and "Down" (rollback) sections
- Goose tracks which migrations ran
- Run on app startup

### Step 4.4: Build Store Implementations

**For each store, implement:**

**PostgresWorkoutStore:**

- CRUD operations with SQL queries
- Use transactions (workout + entries saved together)
- Handle relationships (JOIN queries)

**PostgresUserStore:**

- User creation with password hashing
- Lookup by username
- Token validation with JOIN

**PostgresTokenStore:**

- Insert tokens (store hash, not plaintext!)
- Delete expired tokens
- Generate new tokens

**Key concepts:**

- Use parameterized queries ($1, $2) to prevent SQL injection
- Always close rows (defer rows.Close())
- Handle sql.ErrNoRows (not found vs error)

---

## 🌐 Phase 5: API Layer (Handlers)

### Step 5.1: Design Handler Structure

**Each handler:**

- Takes store interface (not concrete type!)
- Takes logger for error tracking
- Has constructor function (NewWorkoutHandler)

**Why?**

- Testable (can inject mock store)
- Follows dependency injection pattern

### Step 5.2: Implement User Registration

**UserHandler approach:**

1. Create validation function (check email format, password length)
2. In HandleRegisterUser:
   - Decode JSON body
   - Validate input
   - Hash password with bcrypt
   - Call userStore.CreateUser
   - Return user JSON (exclude password hash)

**Design decision:** Server-side validation is MUST (never trust client)

### Step 5.3: Implement Authentication

**TokenHandler approach:**

1. In HandleCreateToken (login):
   - Get username and password from request
   - Lookup user in database
   - Compare password hash using bcrypt
   - Generate secure random token
   - Save token hash to database
   - Return plaintext token to client

**Security principles:**

- Store hash in database (SHA-256)
- Client never sees hash, only plaintext token
- Token expires after 24 hours
- Can't reverse engineer from database

### Step 5.4: Implement Workout CRUD

**WorkoutHandler approach:**

**Create:**

- Get authenticated user from context
- Attach user.ID to workout (ownership)
- Use transaction to save workout + entries
- Return created workout

**Read:**

- Extract ID from URL
- Fetch workout with entries
- Return JSON

**Update:**

- Verify ownership (user owns this workout?)
- Use pointer fields for partial updates
- Replace all entries
- Return updated workout

**Delete:**

- Verify ownership
- Delete workout (CASCADE removes entries)
- Return 204 No Content

**Key pattern:** Check authorization BEFORE database operation!

---

## 🔐 Phase 6: Authentication Middleware

### Step 6.1: Design Context System

**Problem:** How to pass user between middleware and handlers?

**Solution:** Use Go's context!

**SetUser function:**

- Takes request and user
- Adds user to request context
- Returns modified request

**GetUser function:**

- Extracts user from context
- Returns user to handler

### Step 6.2: Build Authenticate Middleware

**Purpose:** Validate token and set user in context

**Flow:**

1. Check Authorization header
2. If empty → set AnonymousUser
3. Parse "Bearer TOKEN" format
4. Hash token with SHA-256
5. Lookup in database (check expiry)
6. Fetch associated user
7. Inject user into context
8. Call next handler

**Key:** Always call next handler (with anonymous or real user)

### Step 6.3: Build RequireUser Middleware

**Purpose:** Block anonymous users from protected routes

**Flow:**

1. Get user from context
2. Check if IsAnonymousUser()
3. If yes → return 401 Unauthorized
4. If no → allow request to continue

**Usage:** Wrap protected routes with this middleware

---

## 🛣️ Phase 7: Route Configuration

### Step 7.1: Set Up Chi Router

**Why Chi?**

- Lightweight and fast
- Middleware support
- URL parameters easy to extract

### Step 7.2: Configure Route Groups

**Organize routes by access level:**

**Public routes (no auth needed):**

- POST /users (registration)
- POST /tokens/authentication (login)
- GET /health (health check)

**Protected routes (requires token):**

- Use r.Group with Authenticate middleware
- Wrap handlers with RequireUser
- All /workouts endpoints

**Pattern:**

```
Router
  ├─ Public routes (anyone can access)
  └─ Group with authentication
       └─ Protected routes (must be logged in)
```

---

## 🛠️ Phase 8: Utilities & Helpers

### Step 8.1: JSON Response Helper

**Problem:** Writing JSON responses is repetitive

**Solution:** Create WriteJson function

- Takes status code and data envelope
- Pretty prints JSON (easier debugging)
- Sets correct content type
- Handles errors centrally

### Step 8.2: URL Parameter Helper

**Problem:** Extracting and validating IDs from URLs is repetitive

**Solution:** Create ReadIDParam function

- Extracts {id} from URL
- Converts string to int64
- Validates format
- Returns error if invalid

---

## 🧪 Phase 9: Testing Setup (Optional but Recommended)

### Step 9.1: Design Test Database

**Approach:**

- Separate database for tests (port 5500)
- Fresh migrations before each test
- Truncate tables after each test

### Step 9.2: Write Store Tests

**Pattern:**

- Create setupTestDB helper
- Test each CRUD operation
- Use real database (integration tests)
- Verify data with assertions

**Why test stores?**

- Ensures SQL queries work
- Catches migration issues
- Documents expected behavior

---

## 🚀 Phase 10: Development Workflow

### Step 10.1: Set Up Hot Reload

**Use Air:**

- Watches .go files for changes
- Automatically rebuilds
- Restarts server

**Configuration:**

- Create .air.toml (optional)
- Define what to watch
- Set build command

### Step 10.2: Database Management Commands

**Keep handy:**

- Start DB: `docker compose up -d`
- Stop DB: `docker compose down`
- Run migrations: `goose -dir migrations postgres "connection-string" up`
- Check status: `goose -dir migrations postgres "connection-string" status`

### Step 10.3: Testing with Curl/Postman

**Build test script:**

1. Register user
2. Login to get token
3. Create workout with token
4. Update workout
5. Delete workout

**Save examples in post.md or similar**

---

## 🎨 Design Patterns Used

### 1. Repository Pattern (Store Layer)

**What:** Separate data access from business logic
**Why:** Easy to test, swap databases, maintain

### 2. Dependency Injection

**What:** Pass dependencies through constructors
**Why:** Explicit, testable, flexible

### 3. Middleware Chain

**What:** Sequential request processing
**Why:** Reusable logic (auth, logging), clean separation

### 4. Interface-Based Design

**What:** Define contracts, code to interfaces
**Why:** Loose coupling, mockable, swappable implementations

### 5. Transaction Management

**What:** Group related DB operations
**Why:** Data consistency (all or nothing)

### 6. Context for Request-Scoped Data

**What:** Pass user through request context
**Why:** Type-safe, doesn't pollute function signatures

---

## 🔄 Development Order (Recommended)

### Stage 1: Foundation (Do First)

1. Project structure setup
2. Database connection
3. Migrations for user table
4. Docker setup

### Stage 2: Authentication (Do Second)

5. User store implementation
6. User handler (registration)
7. Token generation logic
8. Token store implementation
9. Token handler (login)
10. Authentication middleware

### Stage 3: Core Feature (Do Third)

11. Workout store implementation
12. Workout handler (CRUD)
13. Authorization checks (ownership)
14. Protected routes setup

### Stage 4: Polish (Do Last)

15. Error handling improvements
16. Validation enhancements
17. Testing
18. Documentation

**Why this order?**

- Can test each piece independently
- Authentication needed before protected features
- Build on solid foundation

---

## 🎯 Key Principles Throughout

### 1. Security First

- Hash passwords (bcrypt)
- Hash tokens (SHA-256)
- Validate all input
- Check authorization
- Use parameterized queries

### 2. Error Handling

- Log errors with context
- Return user-friendly messages
- Don't leak sensitive info
- Use proper HTTP status codes

### 3. Code Organization

- One responsibility per function
- Keep handlers thin
- Put business logic in stores
- Use meaningful names

### 4. Database Best Practices

- Always use transactions for related operations
- Close resources (defer rows.Close())
- Handle NULL values with pointers
- Use foreign keys and CASCADE

### 5. API Design

- RESTful endpoints (/workouts, /users)
- Consistent JSON structure (envelope pattern)
- Proper status codes (200, 201, 204, 400, 401, 403, 500)
- Clear error messages

---

## 🐛 Common Pitfalls to Avoid

### 1. Not Closing Database Resources

**Problem:** Memory leaks from unclosed rows
**Solution:** Always `defer rows.Close()`

### 2. Storing Plaintext Passwords/Tokens

**Problem:** Security nightmare if database leaks
**Solution:** Always hash sensitive data

### 3. Missing Authorization Checks

**Problem:** Users can modify others' data
**Solution:** Always verify ownership before update/delete

### 4. Ignoring Errors

**Problem:** Silent failures, hard to debug
**Solution:** Check every error, log with context

### 5. Not Using Transactions

**Problem:** Partial saves (workout saved, entries failed)
**Solution:** Group related operations in transactions

### 6. Hardcoding Values

**Problem:** Can't change without recompiling
**Solution:** Use flags, environment variables, config files

### 7. Not Validating Input

**Problem:** Bad data in database
**Solution:** Server-side validation always

### 8. Copying Context Values

**Problem:** Race conditions, unexpected behavior
**Solution:** Context is immutable, always use properly

---

## 📚 Learning Resources Used

### Go Fundamentals

- Structs and methods
- Interfaces and polymorphism
- Error handling patterns
- Context package
- Database/sql package

### Web Development

- HTTP methods (GET, POST, PUT, DELETE)
- Status codes
- Headers (Authorization, Content-Type)
- JSON encoding/decoding
- Middleware pattern

### Database

- SQL queries (SELECT, INSERT, UPDATE, DELETE)
- JOINs (INNER JOIN)
- Transactions (BEGIN, COMMIT, ROLLBACK)
- Foreign keys and CASCADE
- Indexes for performance

### Security

- Password hashing (bcrypt)
- Token generation (crypto/rand)
- SHA-256 hashing
- Bearer token authentication
- Authorization vs Authentication

---

## 🎓 Skills Gained from This Project

### Go Programming

✅ Project structure and organization
✅ Interface-based design
✅ Dependency injection
✅ Error handling patterns
✅ Working with database/sql
✅ Middleware implementation
✅ Context usage

### Backend Development

✅ RESTful API design
✅ Authentication systems
✅ Authorization patterns
✅ CRUD operations
✅ Transaction management
✅ Request validation

### Database

✅ Schema design
✅ Migrations
✅ Foreign keys
✅ JOIN queries
✅ Transaction patterns

### Tools & DevOps

✅ Docker for local development
✅ Hot reload with Air
✅ Database migrations with Goose
✅ API testing with curl/Postman

---

## 🚀 Next Steps to Enhance Project

### Easy Additions

- [ ] Add pagination to workout list
- [ ] Add search/filter workouts
- [ ] Add user profile endpoint
- [ ] Add password change endpoint
- [ ] Add email validation endpoint

### Medium Difficulty

- [ ] Add refresh tokens (separate from auth tokens)
- [ ] Add rate limiting middleware
- [ ] Add CORS middleware
- [ ] Add request logging middleware
- [ ] Add workout statistics endpoints

### Advanced Features

- [ ] Add role-based access (admin, user)
- [ ] Add workout sharing between users
- [ ] Add workout templates/programs
- [ ] Add file uploads (profile pictures)
- [ ] Add real-time features (WebSockets)
- [ ] Add API versioning (v1, v2)

### Production Readiness

- [ ] Add comprehensive error logging
- [ ] Add metrics/monitoring
- [ ] Add health check with DB status
- [ ] Add graceful shutdown
- [ ] Add configuration management
- [ ] Add API documentation (Swagger/OpenAPI)
- [ ] Add deployment scripts
- [ ] Add CI/CD pipeline

---

## 💡 Final Thoughts

This project teaches you to **think in layers** and **build systematically**.

Start simple → Add complexity gradually → Test as you go → Refactor when needed.

The architecture is flexible enough to scale from personal project to production API.

**Remember:** Good code is not about being clever, it's about being **clear**, **maintainable**, and **secure**.
