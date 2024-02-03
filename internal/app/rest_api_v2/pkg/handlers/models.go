package handlers

// Standard error response. Used in Swagger docs and inside the main error handler.
type SwaggerError struct {
	ErrorID    int    `json:"id"`
	ErrorValue string `json:"message"`
}

// Purely Swagger related object, not used anywhere else in the codebase
type Models_SimpleSuccess struct {
	Message string `json:"message"` // success
}
