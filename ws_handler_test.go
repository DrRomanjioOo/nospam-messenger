package handler

import (
	"net/http"
)

type HealthChecker interface {
	PingPostgres(r *http.Request) error
	PingRedis(r *http.Request) error
}

type HealthHandler struct {
	checker HealthChecker
}

func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

func (h *HealthHandler) Serve(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	body := map[string]string{"status": "ok"}

	if err := h.checker.PingPostgres(r); err != nil {
		status = http.StatusServiceUnavailable
		body["status"] = "degraded"
		body["postgres"] = err.Error()
	}
	if err := h.checker.PingRedis(r); err != nil {
		status = http.StatusServiceUnavailable
		body["status"] = "degraded"
		body["redis"] = err.Error()
	}

	writeJSON(w, status, body)
}
