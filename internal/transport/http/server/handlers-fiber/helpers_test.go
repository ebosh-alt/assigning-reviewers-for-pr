package handlers_fiber

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"assigning-reviewers-for-pr/internal/entities"
	api "assigning-reviewers-for-pr/internal/oapi"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestWriteErrorPREXISTS(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return writeError(c, entities.ErrPRExists)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusConflict, resp.StatusCode)

	var body api.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, api.PREXISTS, body.Error.Code)
	require.Equal(t, "PR id already exists", body.Error.Message)
}

func TestWriteErrorNotFoundMessage(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return writeError(c, entities.ErrUserNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body api.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, api.NOTFOUND, body.Error.Code)
	require.Equal(t, "resource not found", body.Error.Message)
}

func TestWriteErrorReassignConflicts(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected api.ErrorResponse
	}{
		{
			name: "merged",
			err:  entities.ErrPRMerged,
			expected: api.ErrorResponse{Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{Code: api.PRMERGED, Message: "cannot reassign on merged PR"}},
		},
		{
			name: "not_assigned",
			err:  entities.ErrNotAssigned,
			expected: api.ErrorResponse{Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{Code: api.NOTASSIGNED, Message: "reviewer is not assigned to this PR"}},
		},
		{
			name: "no_candidate",
			err:  entities.ErrNoCandidate,
			expected: api.ErrorResponse{Error: struct {
				Code    api.ErrorResponseErrorCode `json:"code"`
				Message string                     `json:"message"`
			}{Code: api.NOCANDIDATE, Message: "no active replacement candidate in team"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				return writeError(c, tt.err)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusConflict, resp.StatusCode)

			var body api.ErrorResponse
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			require.Equal(t, tt.expected.Error.Code, body.Error.Code)
			require.Equal(t, tt.expected.Error.Message, body.Error.Message)
		})
	}
}
