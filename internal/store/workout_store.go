package store

import "database/sql"

// ? - main workout data structure
type Workout struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration_minutes"`
	CaloriesBurned  int            `json:"calories_burned"`
	Entries         []WorkoutEntry `json:"entries"` // ! nested entries for each exercise
}

// ? - individual exercise within a workout
type WorkoutEntry struct {
	ID              int      `json:"id"`
	ExerciseName    string   `json:"exercise_name"`
	Sets            int      `json:"sets"`
	Reps            *int     `json:"reps"` // * pointer so it can be null
	DurationSeconds *int     `json:"duration_seconds"` // * pointer so it can be null
	Weight          *float64 `json:"weight"` // * pointer so it can be null
	Notes           string   `json:"notes"`
	OrderIndex      int      `json:"order_index"`
}

// * holds the db connection for workout operations
type PostgresWorkoutStore struct {
	db *sql.DB
}

// ? - constructor that creates new store instance
func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{db: db}
}


//! collection of methods
type WorkoutStore interface {
	CreateWorkout(*Workout) (*Workout, error)
	GetWorkoutByID(id int64) (*Workout, error)
	UpdateWorkout(*Workout)  error
	DeleteWorkout(id int64)  error
}

func (pg *PostgresWorkoutStore) CreateWorkout(workout *Workout) (*Workout, error) {
	
	// ! starting a transaction so both workout & entries get saved together
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // ? - rolls back if anything fails

	// * inserting main workout data first
	query :=
		`
  INSERT INTO workouts (user_id, title, description, duration_minutes, calories_burned)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING id 
  `

	err = tx.QueryRow(query, workout.UserID, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned).Scan(&workout.ID)
	if err != nil {
		return nil, err
	}

	// ? - now looping through each exercise entry and saving them
	for i := range workout.Entries {
		query := `
    INSERT INTO workout_entries (workout_id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING id
    `
		err = tx.QueryRow(query, workout.ID, workout.Entries[i].ExerciseName, workout.Entries[i].Sets, workout.Entries[i].Reps, workout.Entries[i].DurationSeconds, workout.Entries[i].Weight, workout.Entries[i].Notes, workout.Entries[i].OrderIndex).Scan(&workout.Entries[i].ID)
		if err != nil {
			return nil, err
		}
	}

	// ! commit the transaction - makes everything permanent
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) GetWorkoutByID(id int64) (*Workout, error) {
	workout := &Workout{}
	// * fetching main workout info by id
	query := `
  SELECT id, title, description, duration_minutes, calories_burned
  FROM workouts
  WHERE id = $1
  `
	err := pg.db.QueryRow(query, id).Scan(&workout.ID, &workout.Title, &workout.Description, &workout.DurationMinutes, &workout.CaloriesBurned)
	if err == sql.ErrNoRows {
		return nil, nil // ? - workout doesn't exist
	}

	if err != nil {
		return nil, err
	}

	// ? - now grabbing all exercise entries for this workout
	entryQuery := `
  SELECT id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index
  FROM workout_entries
  WHERE workout_id = $1
  ORDER BY order_index
  `

	rows, err := pg.db.Query(entryQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // ! important - close when done

	// * looping through all the entries
	for rows.Next() {
		var entry WorkoutEntry
		err = rows.Scan(
			&entry.ID,
			&entry.ExerciseName,
			&entry.Sets,
			&entry.Reps,
			&entry.DurationSeconds,
			&entry.Weight,
			&entry.Notes,
			&entry.OrderIndex,
		)
		if err != nil {
			return nil, err
		}
		workout.Entries = append(workout.Entries, entry) // ? - attaching entry to workout
	}

	return workout, nil
}

func (pg *PostgresWorkoutStore) UpdateWorkout(workout *Workout) error {
	// ! transaction for updating both workout and its entries
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // ? - safety net if update fails

	// * updating main workout info
	query := `
  UPDATE workouts
  SET title = $1, description = $2, duration_minutes = $3, calories_burned = $4
  WHERE id = $5
  `

	_, err = tx.Exec(query, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned, workout.ID)
	if err != nil {
		return err
	}

	// ? - wiping old entries first
	_, err = tx.Exec("DELETE FROM workout_entries WHERE workout_id = $1", workout.ID)
	if err != nil {
		return err
	}

	// ? - inserting fresh entries
	for _, entry := range workout.Entries {
		query := `
    INSERT INTO workout_entries (workout_id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
		_, err = tx.Exec(query, workout.ID, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.Notes, entry.OrderIndex)
		if err != nil {
			return err
		}
	}

	// ! commit to save all changes
	return tx.Commit()
}

func (pg *PostgresWorkoutStore) DeleteWorkout(id int64) error {
	//  query for delete
	query := `
	DELETE FROM workouts
	where id=$1
	`

	result,err := pg.db.Exec(query,id)
	if err!= nil {
		return err
	}
	// if any row is affected from failed operation
	rowsAffected,err := result.RowsAffected()
		if err!= nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil

}