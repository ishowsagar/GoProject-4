package api

import (
	"database/sql"
	"encoding/json"
	"fem/internal/store"
	"fem/internal/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// types declaration
type WorkoutHandler struct {
	workstore store.WorkoutStore
	logger *log.Logger

}

func NewWorkoutHandler(workoutStore store.WorkoutStore,logger *log.Logger) *WorkoutHandler {
return &WorkoutHandler{
	workstore: workoutStore,
	logger: logger,
}
}

//! methods --> have base method WorkoutHandler ( points to type which persists changes across app) --> other called via base this one	
func (wh *WorkoutHandler) HandleWorkoutByID(w http.ResponseWriter, req *http.Request) {
// extracting id from url via chi

workoutID,err := utils.ReadIDParam(req)
if err != nil {
	wh.logger.Printf("Error : readIdParam : %v ",err)
	utils.WriteJson(w,http.StatusBadRequest,utils.Envelope{"error" : "Invalid workout id"})
	return
}

workout,err := wh.workstore.GetWorkoutByID(workoutID)
if err != nil {
	// ? - db error fetching workout
	wh.logger.Printf("Error : getWorkoutByID : %v ",err)
	utils.WriteJson(w,http.StatusInternalServerError,utils.Envelope{"error" : "Internal Server Error"})
	return
}
// * sending json response with helper function
utils.WriteJson(w,http.StatusOK,utils.Envelope{"workout":workout})
}


// ! CreateWorkout Method
func (wh *WorkoutHandler) HandleCreateWorkout (w http.ResponseWriter, req *http.Request) {
var workout  store.Workout // * follows type def of this struct
err := json.NewDecoder(req.Body).Decode(&workout)

if err !=nil {
wh.logger.Printf("Error : decodingCreateWorkout : %v ",err)
utils.WriteJson(w,http.StatusBadRequest,utils.Envelope{"error" : "Invalid request sent"})
	return
}

createWorkout,err := wh.workstore.CreateWorkout(&workout)
if err !=nil {
	wh.logger.Printf("Error : createWorkout : %v ",err)
	utils.WriteJson(w,http.StatusInternalServerError,utils.Envelope{"error" : "failed to create workout"})
	return
}

utils.WriteJson(w,http.StatusCreated,utils.Envelope{"workout" : createWorkout})
}

// ! UpdateWorkout Method
func (wh *WorkoutHandler) HandleUpdateWorkoutByID(w http.ResponseWriter,req *http.Request) {
	// extracting id from url via chi

workoutID,err := utils.ReadIDParam(req)
if err!= nil {
	wh.logger.Printf("Error : readIdParam : %v ",err)
	utils.WriteJson(w,http.StatusBadRequest,utils.Envelope{"error" : "Invalid workout update id"})
}
existingWorkout,err := wh.workstore.GetWorkoutByID(workoutID)
if err != nil {
	// ? - db error while fetching workout
	wh.logger.Printf("Error : getWorkoutByID : %v ",err)
	utils.WriteJson(w,http.StatusInternalServerError,utils.Envelope{"error" : "Internal server error"})
	return
}
if existingWorkout == nil {
	http.NotFound(w,req)
	return
}

// found existing workout

var updateWorkoutRequest struct {
		Title           *string              `json:"title"`
		Description     *string              `json:"description"`
		DurationMinutes *int                 `json:"duration_minutes"`
		CaloriesBurned  *int                 `json:"calories_burned"`
		Entries         []store.WorkoutEntry `json:"entries"`
	}
	err = json.NewDecoder(req.Body).Decode(&updateWorkoutRequest) // this body refrences to instance of the struct which persists changes

	if err != nil {
		// ? - failed to parse JSON body
		wh.logger.Printf("Error : decodingUpdateRequest : %v ",err)
	    utils.WriteJson(w,http.StatusBadRequest,utils.Envelope{"error" : "Invalid request payload"})
		return
	}

	// if coming body has title available
	if updateWorkoutRequest.Title != nil {
		existingWorkout.Title = *updateWorkoutRequest.Title
	}

	if updateWorkoutRequest.Description != nil {
		existingWorkout.Description = *updateWorkoutRequest.Description
	}
	if updateWorkoutRequest.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updateWorkoutRequest.DurationMinutes
	}
	if updateWorkoutRequest.CaloriesBurned != nil {
		existingWorkout.CaloriesBurned = *updateWorkoutRequest.CaloriesBurned
	}
	if updateWorkoutRequest.Entries != nil {
		existingWorkout.Entries = updateWorkoutRequest.Entries
	}

	// ! make sure ID is set for the update
	existingWorkout.ID = int(workoutID)
	
	err = wh.workstore.UpdateWorkout(existingWorkout)
	if err !=nil {
		// ? - db error while updating
		wh.logger.Printf("Error : updateWorkout : %v ",err)
		utils.WriteJson(w,http.StatusInternalServerError,utils.Envelope{"error" : "Internal server error"})
		return
	}

	// * sending response
	utils.WriteJson(w,http.StatusOK,utils.Envelope{"workout":existingWorkout})
}

func (wh *WorkoutHandler) HandleDeleteWorkoutByID(w http.ResponseWriter, req *http.Request)  {
	// extracting id from url via chi
paramsWorkoutID := chi.URLParam(req,"id") // passing req and "slug" being route params

// if id not found on url params
if paramsWorkoutID == "" {
	http.NotFound(w,req)
	return
}

workoutID,err := strconv.ParseInt(paramsWorkoutID,10,64)
if err != nil {
http.NotFound(w,req)
return
}  

err = wh.workstore.DeleteWorkout(workoutID)
if err == sql.ErrNoRows {
http.Error(w,"Workout not found",http.StatusNotFound)
return
}  
if err != nil {
http.Error(w,"error deleting the workout", http.StatusInternalServerError)
return
}  

w.WriteHeader(http.StatusNoContent)
}

// ! WORKFLOW BREAKDOWN:
// 1. WorkoutStore interface --> defines what methods our store needs (contract/blueprint)
// 2. PostgresWorkoutStore --> actual implementation with real SQL queries
// 3. WorkoutHandler holds workstore (the interface) --> can swap db types easily
// 4. Handler methods (like HandleCreateWorkout) --> call wh.workstore.CreateWorkout() 
// 5. Routes hook these handlers to URLs --> /workouts maps to HandleCreateWorkout
// ? - interface lets us swap PostgreSQL for MySQL/MongoDB without touching handlers! 