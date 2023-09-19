package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// writeJSON is a helper which encodes the given data to JSON, writes it with
// a 200 OK status code, and sets the Content-Type header to application/json.
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil
}

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
	w.WriteHeader(http.StatusBadRequest)
	w.Write(out)

	return nil
}

// unauthorized is a helper to send a JSON response with a 401 Unauthorized
// status code and a given error message.
func (app *application) unauthorized(w http.ResponseWriter) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = "Unauthorized"

	err := app.writeJSON(w, http.StatusUnauthorized, payload)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

// passwordMatches is a helper which takes a bcrypt hash and a password as
// inputs, and returns true if the password matches the hash. Otherwise it
// returns false. It also returns an error if there was an issue with the
// bcrypt package (if the hash isn't in the expected format).
func (app *application) passwordMatches(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
