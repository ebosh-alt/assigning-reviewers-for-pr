package handlers_fiber

import (
	"net/http"

	"assigning-reviewers-for-pr/internal/entities"
	api "assigning-reviewers-for-pr/internal/oapi"

	"github.com/gofiber/fiber/v2"
)

// GetStats returns базовую агрегацию.
func (h *Handler) GetStats(c *fiber.Ctx) error {
	statsRes, err := h.uc.Stats(c.Context())
	if err != nil {
		h.log.Errorw("failed to get stats", "error", err.Error())
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(statsRes)
}

// GetStatsSummary возвращает отфильтрованную статистику.
func (h *Handler) GetStatsSummary(c *fiber.Ctx, params api.GetStatsSummaryParams) error {
	filter := entities.StatsFilter{}
	if params.From != nil {
		filter.From = params.From
	}
	if params.To != nil {
		filter.To = params.To
	}
	if params.Status != nil {
		status := entities.PullRequestStatus(*params.Status)
		filter.Status = &status
	}
	if params.Limit != nil && *params.Limit > 0 {
		filter.Limit = int(*params.Limit)
	}

	summary, err := h.uc.SummaryStats(c.Context(), filter)
	if err != nil {
		h.log.Errorw("failed to get summary stats", "error", err.Error())
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(summary)
}

// GetStatsReviewerUserId возвращает статистику ревьюера.
func (h *Handler) GetStatsReviewerUserId(c *fiber.Ctx, userID string, params api.GetStatsReviewerUserIdParams) error {
	limit := 10
	if params.Limit != nil && *params.Limit > 0 {
		limit = int(*params.Limit)
	}

	res, err := h.uc.ReviewerStats(c.Context(), userID, limit)
	if err != nil {
		h.log.Errorw("failed to get reviewer stats", "error", err.Error())
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(res)
}

// GetStatsPrPrId возвращает статистику по PR.
func (h *Handler) GetStatsPrPrId(c *fiber.Ctx, prID string) error {
	res, err := h.uc.PRStats(c.Context(), prID)
	if err != nil {
		h.log.Errorw("failed to get PR stats", "error", err.Error())
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(res)
}
