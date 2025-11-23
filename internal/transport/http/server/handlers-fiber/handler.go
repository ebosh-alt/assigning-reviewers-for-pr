// Package handlers_fiber wires HTTP delivery components.
package handlers_fiber

import (
	"assigning-reviewers-for-pr/internal/usecase"
	"go.uber.org/zap"
)

// Handler implements oapi.ServerInterface using service layer interfaces.
type Handler struct {
	log *zap.SugaredLogger
	uc  usecase.InterfaceUsecase
}

// NewHandler constructs an HTTP server with service dependencies.
func NewHandler(log *zap.SugaredLogger, usecase usecase.InterfaceUsecase) *Handler {
	return &Handler{
		log: log,
		uc:  usecase,
	}
}
