package handlers_fiber

import (
	"net/http"

	"assigning-reviewers-for-pr/internal/mapper"
	api "assigning-reviewers-for-pr/internal/oapi"
	"github.com/gofiber/fiber/v2"
)

// GetUsersGetReview returns PRs where user is reviewer.
func (h *Handler) GetUsersGetReview(c *fiber.Ctx, params api.GetUsersGetReviewParams) error {
	prs, err := h.uc.GetReviewList(c.Context(), params.UserId)
	if err != nil {
		h.log.Errorw("failed to get review list", "error", err.Error())
		return writeError(c, err)
	}

	resp := struct {
		UserID       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       params.UserId,
		PullRequests: mapper.ToOAPIPullShortList(prs),
	}

	return c.Status(http.StatusOK).JSON(resp)
}

// PostUsersSetIsActive toggles user activity flag.
func (h *Handler) PostUsersSetIsActive(c *fiber.Ctx) error {
	var body api.PostUsersSetIsActiveJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		h.log.Errorw("failed to parse body", "error", err.Error())
		return c.Status(http.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "invalid body"))
	}

	usr, err := h.uc.SetActiveUser(c.Context(), body.UserId, body.IsActive)
	if err != nil {
		h.log.Errorw("failed to set is_active for user", "error", err.Error())
		return writeError(c, err)
	}

	resp := struct {
		User api.User `json:"user"`
	}{User: mapper.ToOAPIUser(*usr)}
	return c.Status(http.StatusOK).JSON(resp)
}
