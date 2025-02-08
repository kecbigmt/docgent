package application

import (
	"net/http"

	"go.uber.org/zap"
)

type HealthHandler struct {
	log *zap.Logger
}

func NewHealthHandler(log *zap.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

func (h *HealthHandler) Pattern() string {
	return "/health"
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
