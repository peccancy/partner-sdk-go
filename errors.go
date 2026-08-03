package partner

// APIError is returned when the API responds with a non-2xx status (or the request fails).
type APIError struct {
	StatusCode int
	Body       string
	message    string
}

func (e *APIError) Error() string { return e.message }
