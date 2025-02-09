package lib

import "fmt"

type HTTPError struct {
	StatusCode int
	Status     string
	RawBody    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: %s", e.Status)
}
