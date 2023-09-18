package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// readJSON is a helper which decodes the JSON request body into the interface
// provided as a parameter.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1_048_576 // 1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		app.errorLog.Println("ERROR to decoded JSON: ", err)
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// badRequest is a helper to send a JSON response with a 400 Bad Request status
// code and a given error message.
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)

	return nil
}
