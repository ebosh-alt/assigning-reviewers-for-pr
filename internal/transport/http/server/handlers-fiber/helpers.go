package handlers_fiber

import (
	"errors"
	"net/http"

	"assigning-reviewers-for-pr/internal/entities"
	api "assigning-reviewers-for-pr/internal/oapi"
	"github.com/gofiber/fiber/v2"
)

func writeError(c *fiber.Ctx, err error) error {
	status := http.StatusInternalServerError
	code := api.NOTFOUND
	msg := "internal error"

	switch {
	case errors.Is(err, entities.ErrInvalidArgument):
		status = http.StatusBadRequest
		code = api.NOTFOUND
		msg = err.Error()
	case errors.Is(err, entities.ErrUserNotFound), errors.Is(err, entities.ErrTeamNotFound), errors.Is(err, entities.ErrPRNotFound):
		status = http.StatusNotFound
		code = api.NOTFOUND
		msg = "resource not found"
	case errors.Is(err, entities.ErrTeamExists):
		status = http.StatusBadRequest
		code = api.TEAMEXISTS
		msg = "team_name already exists"
	case errors.Is(err, entities.ErrPRExists):
		status = http.StatusConflict
		code = api.PREXISTS
		msg = "PR id already exists"
	case errors.Is(err, entities.ErrPRMerged):
		status = http.StatusConflict
		code = api.PRMERGED
		msg = "cannot reassign on merged PR"
	case errors.Is(err, entities.ErrNotAssigned):
		status = http.StatusConflict
		code = api.NOTASSIGNED
		msg = "reviewer is not assigned to this PR"
	case errors.Is(err, entities.ErrNoCandidate):
		status = http.StatusConflict
		code = api.NOCANDIDATE
		msg = "no active replacement candidate in team"
	default:
		msg = err.Error()
	}

	return c.Status(status).JSON(errorResponse(code, msg))
}

func errorResponse(code api.ErrorResponseErrorCode, msg string) api.ErrorResponse {
	return api.ErrorResponse{Error: struct {
		Code    api.ErrorResponseErrorCode `json:"code"`
		Message string                     `json:"message"`
	}{Code: code, Message: msg}}
}
