package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// type declaration
type Envelope map[string]interface{}

// util function to be exported out for use
func WriteJson(w http.ResponseWriter, status int, data Envelope) error {
	// ? - using MarshalIndent with "\t" for pretty formatted JSON output
	json,err := json.MarshalIndent(data,""," ")
	
	if err != nil {
		return err
	}

	json = append(json, '\n')
	w.Header().Set("Content-type","application/json")
	w.WriteHeader(status)
	w.Write(json)
	return nil
}

// 
func ReadIDParam(r *http.Request) (int64,error) {
	idParam := chi.URLParam(r,"id") // reads id slug from the client's request url
	
	// if no id was passed
	if idParam == "" {
		return 0, errors.New("Invalid id parameter")
	}
	id,err := strconv.ParseInt(idParam,10,64)
	if err!= nil {
		return 0, errors.New("Invalid id parameter type")
	}
	
	return id,err
}