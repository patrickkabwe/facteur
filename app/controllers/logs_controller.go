package controllers

import (
	"citadel/app/drivers"
	"citadel/app/models"
	"citadel/app/repositories"
	appsPages "citadel/views/pages/apps"

	"github.com/caesar-rocks/auth"
	caesar "github.com/caesar-rocks/core"
)

type LogsController struct {
	driver   drivers.Driver
	appsRepo *repositories.ApplicationsRepository
}

func NewLogsController(driver drivers.Driver, appsRepo *repositories.ApplicationsRepository) *LogsController {
	return &LogsController{driver, appsRepo}
}

func (c *LogsController) Index(ctx *caesar.CaesarCtx) error {
	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	if err != nil {
		return err
	}

	app, err := c.appsRepo.FindOneBy(ctx.Context(), "slug", ctx.PathValue("slug"))
	if err != nil {
		return caesar.NewError(400)
	}

	if app.UserID != user.ID {
		return caesar.NewError(403)
	}

	return ctx.Render(appsPages.LogsPage(*app))
}

func (c *LogsController) Stream(ctx *caesar.CaesarCtx) error {
	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	if err != nil {
		return err
	}

	app, err := c.appsRepo.FindOneBy(ctx.Context(), "slug", ctx.PathValue("slug"))
	if err != nil {
		return err
	}

	if app.UserID != user.ID {
		return caesar.NewError(403)
	}

	ctx.SetSSEHeaders()

	return c.driver.StreamLogs(ctx, *app)
}
