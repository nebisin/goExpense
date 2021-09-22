package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit the size of the request body to 1MB
	maxBytes := 1_038_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// decode the request body
	err = json.Unmarshal(body, dst)
	// if there is an error during decoding:
	if err != nil {
		// There is a syntax problem with the JSON
		var syntaxError *json.SyntaxError
		// JSON value is not appropriate for the destination Go type
		var unmarshalTypeError *json.UnmarshalTypeError
		// The decode destination is not valid
		// usually because it is not a pointer.
		// This is a problem with our application.
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			// check if error has the type json.SyntaxError
			// if it does, return a plain error message
			// which includes the location of problem.
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			// this occurs when the wrong type for the target destination
			// if the error relates to a specific field include that in the message
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			// If the request body is empty:
			return errors.New("body must not be empty")

		case err.Error() == "http: request body too large":
			// If request body exceeds 1MB in size the decode will fail.
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			// this will happen if we pass a non-nil pointer to Decode()
			// we catch this and panic instead of returning the error
			// because this error should not happen during normal operation
			// and probably the result of a developer mistake
			panic(err)

		// For anything else return the error message as it is:
		default:
			return err
		}
	}

	return nil
}
