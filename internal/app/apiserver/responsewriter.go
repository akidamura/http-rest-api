package apiserver

import "net/http"

// responseWriter оборачивает http.ResponseWriter,
// чтобы запомнить статус ответа для логирования.
type responseWriter struct {
	http.ResponseWriter
	code int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
