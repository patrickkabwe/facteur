package controllers

import (
	"citadel/app/drivers"
	"citadel/app/models"
	"citadel/app/repositories"
	"citadel/app/services"
	"citadel/app/toast"
	appsPages "citadel/views/pages/apps"

	"github.com/caesar-rocks/auth"
	caesar "github.com/caesar-rocks/core"
	"github.com/charmbracelet/log"
	"github.com/rs/xid"
)

type AppsController struct {
	appsService *services.AppsService
	repo        *repositories.ApplicationsRepository
	driver      drivers.Driver
}

func NewAppsController(appsService *services.AppsService, repo *repositories.ApplicationsRepository, driver drivers.Driver) *AppsController {
	return &AppsController{appsService, repo, driver}
}

func (c *AppsController) Index(ctx *caesar.CaesarCtx) error {
	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	if err != nil {
		return err
	}

	apps, err := c.repo.FindAllFromUser(ctx.Context(), user.ID)
	if err != nil {
		log.Error("err", err)
		return caesar.NewError(400)
	}

	if ctx.WantsJSON() {
		return ctx.SendJSON(apps)
	}

	return ctx.Render(appsPages.IndexPage(apps))
}

type StoreAppValidator struct {
	Name                 string `form:"name" validate:"required,min=3,lowercase"`
	CpuConfig            string `form:"cpu_config" validate:"required"`
	RamConfig            string `form:"ram_config" validate:"required"`
	GitHubInstallationID int64  `form:"github_installation_id"`
	GitHubRepository     string `form:"github_repository"`
	GitHubBranch         string `form:"github_branch"`
}

func (c *AppsController) Store(ctx *caesar.CaesarCtx) error {
	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	if err != nil {
		return err
	}

	data, _, ok := caesar.Validate[StoreAppValidator](ctx)
	if !ok {
		return ctx.Redirect("/apps")
	}

	if data.CpuConfig == "" {
		data.CpuConfig = "shared-cpu-1x"
	}
	if data.RamConfig == "" {
		data.RamConfig = "256MB"
	}

	app := &models.Application{
		ID:                   xid.New().String(),
		UserID:               user.ID,
		Name:                 data.Name,
		CpuConfig:            data.CpuConfig,
		RamConfig:            data.RamConfig,
		GitHubInstallationID: data.GitHubInstallationID,
		GitHubRepository:     data.GitHubRepository,
		GitHubBranch:         data.GitHubBranch,
	}
	if err := c.repo.Create(ctx.Context(), app); err != nil {
		return caesar.NewError(400)
	}

	if err := c.driver.CreateApplication(*app); err != nil {
		return caesar.NewError(400)
	}

	if ctx.WantsJSON() {
		return ctx.SendJSON(app)
	}

	return ctx.Redirect("/apps/" + app.Slug)
}

func (c *AppsController) Show(ctx *caesar.CaesarCtx) error {
	app, err := c.appsService.GetAppOwnedByCurrentUser(ctx)
	if err != nil {
		return err
	}

	return ctx.Render(appsPages.ShowPage(*app))
}

func (c *AppsController) Edit(ctx *caesar.CaesarCtx) error {
	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	if err != nil {
		return err
	}

	app, err := c.repo.FindOneBy(ctx.Context(), "slug", ctx.PathValue("slug"))
	if err != nil {
		return err
	}

	if app.UserID != user.ID {
		return caesar.NewError(403)
	}

	return ctx.Render(appsPages.EditPage(*app))
}

type UpdateApplicationValidator struct {
	Name           string `form:"name" validate:"required,min=3"`
	ReleaseCommand string `form:"release_command"`
	CpuConfig      string `form:"cpu_config" validate:"required"`
	RamConfig      string `form:"ram_config" validate:"required"`
}

func (c *AppsController) Update(ctx *caesar.CaesarCtx) error {
	app, err := c.appsService.GetAppOwnedByCurrentUser(ctx)
	if err != nil {
		return err
	}

	data, errors, ok := caesar.Validate[UpdateApplicationValidator](ctx)
	if !ok {
		return ctx.Render(appsPages.ApplicationsSettingsForm(*app, errors))
	}

	app.Name = data.Name
	app.ReleaseCommand = data.ReleaseCommand
	app.CpuConfig = data.CpuConfig
	app.RamConfig = data.RamConfig

	if err := c.repo.UpdateOneWhere(ctx.Context(), "slug", ctx.PathValue("slug"), app); err != nil {
		return err
	}

	toast.Success(ctx, "Application updated successfully.")

	return ctx.Render(appsPages.ApplicationsSettingsForm(*app, nil))
}

type ConnectGitHubValidator struct {
	GitHubInstallationID int64  `form:"github_installation_id" validate:"required"`
	GitHubRepository     string `form:"github_repository" validate:"required"`
	GitHubBranch         string `form:"github_branch" validate:"required"`
}

func (c *AppsController) ConnectGitHub(ctx *caesar.CaesarCtx) error {
	app, err := c.appsService.GetAppOwnedByCurrentUser(ctx)
	if err != nil {
		return err
	}

	data, _, ok := caesar.Validate[ConnectGitHubValidator](ctx)
	if !ok {
		return ctx.Render(appsPages.EditPage(*app))
	}

	app.GitHubInstallationID = data.GitHubInstallationID
	app.GitHubRepository = data.GitHubRepository
	app.GitHubBranch = data.GitHubBranch

	if err := c.repo.UpdateOneWhere(ctx.Context(), "slug", ctx.PathValue("slug"), app); err != nil {
		return err
	}

	return ctx.Redirect("/apps/" + app.Slug + "/edit")
}

func (c *AppsController) DisconnectGitHub(ctx *caesar.CaesarCtx) error {
	app, err := c.appsService.GetAppOwnedByCurrentUser(ctx)
	if err != nil {
		return err
	}

	app.GitHubInstallationID = -1
	app.GitHubRepository = ""
	app.GitHubBranch = ""

	if err := c.repo.UpdateOneWhere(ctx.Context(), "slug", ctx.PathValue("slug"), app); err != nil {
		return err
	}

	return ctx.Redirect("/apps/" + app.Slug + "/edit")
}

func (c *AppsController) Delete(ctx *caesar.CaesarCtx) error {
	app, err := c.appsService.GetAppOwnedByCurrentUser(ctx)
	if err != nil {
		return err
	}

	if err := c.repo.DeleteOneWhere(ctx.Context(), "id", app.ID); err != nil {
		return err
	}

	return ctx.Redirect("/apps")
}
