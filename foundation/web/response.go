package web

import (
	"context"
	"encoding/json"
	"net/http"
)

// Respond converts a Go value to JSON and sends it to the client.
//
// Isn't this a business concern, thus live in the business layer?
// If we decide the web package is setting a policy for how we communicate, it creates more consistency to put in here.
// If we decide to give teams the flexibility to use other return types; this goes in business layer.
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {

	// Set the status code for the request logger middleware
	SetStatusCode(ctx, statusCode)

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
