package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// types declaration
type WorkoutHandler struct {

}

func NewWorkoutHandler() *WorkoutHandler {
return &WorkoutHandler{}
}

//! methods --> have base method WorkoutHandler ( points to type which persists changes across app) --> other called via base this one	
func (wh *WorkoutHandler) HandleWorkoutByID(w http.ResponseWriter, req *http.Request) {
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
fmt.Fprintf(w,"this is the workout id : %d\n",workoutID)
}

func (wh *WorkoutHandler) HandleCreateWorkout (w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w,"create a workout")

}