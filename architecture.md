# Go Workout API Architecture

## Application Flow Diagram

```mermaid
flowchart TD
    Start([Application Start]) --> ParseFlags[Parse Command-Line Flags<br/>Port: 8080 default]
    ParseFlags --> InitApp[Initialize Application<br/>app.NewApplication]

    InitApp --> ConnectDB[Connect to PostgreSQL<br/>store.Open]
    ConnectDB --> DBCheck{Database<br/>Connection<br/>Success?}
    DBCheck -->|No| Panic1[Panic: DB Error]
    DBCheck -->|Yes| InitLogger[Initialize Logger<br/>log.New]

    InitLogger --> InitHandler[Initialize WorkoutHandler<br/>api.NewWorkoutHandler]
    InitHandler --> CreateAppStruct[Create Application Struct<br/>Logger + WorkoutHandler + DB]
    CreateAppStruct --> AppCheck{App Init<br/>Success?}

    AppCheck -->|No| Panic2[Panic: App Error]
    AppCheck -->|Yes| DeferClose[Defer DB.Close]
    DeferClose --> SetupRoutes[Setup Chi Router<br/>routes.SetupRoutes]

    SetupRoutes --> Route1[GET /health → HealthCheck]
    SetupRoutes --> Route2[GET /workouts/:id → HandleWorkoutByID]
    SetupRoutes --> Route3[POST /workouts → HandleCreateWorkout]

    Route1 --> CreateServer[Create HTTP Server<br/>Port: 8080<br/>Timeouts & Handler]
    Route2 --> CreateServer
    Route3 --> CreateServer

    CreateServer --> ListenServe[server.ListenAndServe]
    ListenServe --> ServerCheck{Server<br/>Start<br/>Success?}

    ServerCheck -->|No| Fatal[Logger.Fatal]
    ServerCheck -->|Yes| Running([Server Running<br/>Listening on :8080])

    Running --> Request{Incoming<br/>Request}

    Request -->|GET /health| Health[HealthCheck Handler<br/>Returns: Status Available]
    Health --> Response1[HTTP Response]

    Request -->|GET /workouts/:id| GetID[HandleWorkoutByID]
    GetID --> ExtractID[Extract ID from URL<br/>chi.URLParam]
    ExtractID --> ValidateID{Valid ID?}
    ValidateID -->|No| NotFound[404 Not Found]
    ValidateID -->|Yes| ReturnWorkout[Return Workout ID]
    ReturnWorkout --> Response2[HTTP Response]
    NotFound --> Response2

    Request -->|POST /workouts| CreateWorkout[HandleCreateWorkout]
    CreateWorkout --> ProcessCreate[Process Workout Creation]
    ProcessCreate --> Response3[HTTP Response]

    Response1 --> Request
    Response2 --> Request
    Response3 --> Request

    style Start fill:#90EE90
    style Running fill:#87CEEB
    style Panic1 fill:#FFB6C1
    style Panic2 fill:#FFB6C1
    style Fatal fill:#FFB6C1
    style ConnectDB fill:#FFE4B5
    style CreateServer fill:#DDA0DD
    style Request fill:#F0E68C
```

## Project Structure

```
GoProject#4/
├── main.go                          # Entry point - server setup
├── go.mod                           # Dependencies
├── internal/
│   ├── app/
│   │   └── app.go                   # Application struct & initialization
│   ├── api/
│   │   └── workout_handler.go       # HTTP handlers for workouts
│   ├── routes/
│   │   └── routes.go                # Chi router configuration
│   └── store/
│       ├── database.go              # PostgreSQL connection
│       └── workout_store.go         # Database operations (empty)
└── tmp/
    └── main.exe~                    # Air build output
```

## API Endpoints

| Method | Path             | Handler             | Description           |
| ------ | ---------------- | ------------------- | --------------------- |
| GET    | `/health`        | HealthCheck         | Health check endpoint |
| GET    | `/workouts/{id}` | HandleWorkoutByID   | Get workout by ID     |
| POST   | `/workouts`      | HandleCreateWorkout | Create new workout    |

## Components

### 1. **Application** (`internal/app/app.go`)

- `Logger` - Logging to stdout
- `WorkoutHandler` - Request handlers
- `DB` - PostgreSQL connection

### 2. **Database** (`internal/store/database.go`)

- Driver: `pgx/v4/stdlib`
- Connection: `localhost:5432`
- Database: `postgres`

### 3. **Router** (`internal/routes/routes.go`)

- Framework: Chi router
- Routes: Health check & workout CRUD operations

### 4. **Handlers** (`internal/api/workout_handler.go`)

- Extract URL parameters with Chi
- Validate input
- Return appropriate HTTP responses
