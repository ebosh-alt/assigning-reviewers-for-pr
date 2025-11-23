package handlers_fiber

import (
	"net/http"

	"assigning-reviewers-for-pr/internal/entities"
	"assigning-reviewers-for-pr/internal/mapper"
	api "assigning-reviewers-for-pr/internal/oapi"
	"github.com/gofiber/fiber/v2"
)

// PostPullRequestCreate handles PR creation with auto assignment.
func (h *Handler) PostPullRequestCreate(c *fiber.Ctx) error {
	var body api.PostPullRequestCreateJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "invalid body"))
	}
	pr, err := h.uc.CreatePullRequest(c.Context(), entities.PullRequest{
		ID:       body.PullRequestId,
		Name:     body.PullRequestName,
		AuthorID: body.AuthorId,
	})
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(http.StatusCreated).JSON(struct {
		PR api.PullRequest `json:"pr"`
	}{PR: mapper.ToOAPIPull(*pr)})
}

// PostPullRequestMerge handles idempotent merge of PR.
func (h *Handler) PostPullRequestMerge(c *fiber.Ctx) error {
	var body api.PostPullRequestMergeJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "invalid body"))
	}
	pr, err := h.uc.MergePullRequest(c.Context(), body.PullRequestId)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(struct {
		PR api.PullRequest `json:"pr"`
	}{PR: mapper.ToOAPIPull(*pr)})
}

// PostPullRequestReassign swaps a reviewer within team.
func (h *Handler) PostPullRequestReassign(c *fiber.Ctx) error {
	var body api.PostPullRequestReassignJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "invalid body"))
	}
	pr, replaced, err := h.uc.ReassignPullRequest(c.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(struct {
		PR         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}{PR: mapper.ToOAPIPull(*pr), ReplacedBy: replaced})
}
