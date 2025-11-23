package handlers_fiber

import (
	"net/http"
	"strings"

	"assigning-reviewers-for-pr/internal/mapper"
	api "assigning-reviewers-for-pr/internal/oapi"
	"github.com/gofiber/fiber/v2"
)

// PostTeamAdd creates a team and upserts members.
func (h *Handler) PostTeamAdd(c *fiber.Ctx) error {
	var body api.PostTeamAddJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "invalid body"))
	}

	team, err := h.uc.CreateTeam(c.Context(), mapper.FromOAPITeam(body))
	if err != nil {
		h.log.Infow(err.Error())
		return writeError(c, err)
	}

	return c.Status(http.StatusCreated).JSON(struct {
		Team api.Team `json:"team"`
	}{Team: mapper.ToOAPITeam(*team)})
}

// GetTeamGet returns team with members by name.
func (h *Handler) GetTeamGet(c *fiber.Ctx, params api.GetTeamGetParams) error {
	team, err := h.uc.Team(c.Context(), params.TeamName)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(http.StatusOK).JSON(mapper.ToOAPITeam(*team))
}

// PostTeamDeactivate деактивирует пользователей команды и переназначает ревьюеров.
func (h *Handler) PostTeamDeactivate(c *fiber.Ctx) error {
	var body api.PostTeamDeactivateJSONRequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "invalid body"))
	}

	teamName := strings.TrimSpace(body.TeamName)
	if teamName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse(api.NOTFOUND, "team_name is required"))
	}

	res, err := h.uc.DeactivateTeam(c.Context(), teamName)
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(res)
}
