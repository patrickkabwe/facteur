package authControllers

import (
	"citadel/app/models"
	authPages "citadel/views/pages/auth"
	"time"

	"github.com/caesar-rocks/auth"
	caesar "github.com/caesar-rocks/core"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
)

type CliController struct {
	redisClient *redis.Client
	auth        *auth.Auth
}

func NewCliController(redisClient *redis.Client, auth *auth.Auth) *CliController {
	return &CliController{redisClient, auth}
}

func (c *CliController) GetSession(ctx *caesar.CaesarCtx) error {
	sessionId := xid.New().String()

	thirtyMinutes := time.Duration(30) * time.Minute

	c.redisClient.Set(ctx.Context(), sessionId, "pending", thirtyMinutes)

	return ctx.SendJSON(map[string]any{
		"sessionId": sessionId,
	})
}

func (c *CliController) Show(ctx *caesar.CaesarCtx) error {
	sessionId := ctx.PathValue("sessionId")
	sessionState := c.redisClient.Get(ctx.Context(), sessionId)
	if sessionState == nil {
		return caesar.NewError(404)
	}

	if sessionState.Val() == "pending" {
		return ctx.Render(authPages.CliPage(false))
	}

	return ctx.Render(authPages.CliPage(true))

}

func (c *CliController) Handle(ctx *caesar.CaesarCtx) error {
	sessionId := ctx.PathValue("sessionId")
	sessionState := c.redisClient.Get(ctx.Context(), sessionId)
	if sessionState == nil {
		return caesar.NewError(404)
	}

	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	if err != nil {
		return err
	}

	jwt, err := c.auth.GenerateJWT(*user)
	if err != nil {
		return err
	}

	if sessionState.Val() != "pending" {
		return caesar.NewError(404)
	}

	fiveMinutes := time.Duration(5) * time.Minute
	c.redisClient.Set(ctx.Context(), sessionId, jwt, fiveMinutes)

	return ctx.Render(authPages.CliSuccessAlert())
}

func (c *CliController) Wait(ctx *caesar.CaesarCtx) error {
	sessionId := ctx.PathValue("sessionId")
	sessionState := c.redisClient.Get(ctx.Context(), sessionId)
	if sessionState == nil {
		return caesar.NewError(404)
	}
	val := sessionState.Val()

	if val == "pending" {
		return ctx.SendJSON(map[string]any{
			"status": "pending",
		})
	}

	c.redisClient.Del(ctx.Context(), sessionId)

	return ctx.SendJSON(map[string]any{
		"status": "done",
		"token":  val,
	})
}

func (c *CliController) Check(ctx *caesar.CaesarCtx) error {
	user, err := auth.RetrieveUserFromCtx[models.User](ctx)
	return ctx.SendJSON(map[string]any{
		"authenticated": err == nil && user != nil,
	})
}
