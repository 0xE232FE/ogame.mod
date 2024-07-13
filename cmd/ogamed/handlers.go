package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alaingilbert/ogame/pkg/gameforge"
	"github.com/alaingilbert/ogame/pkg/wrapper"
	"github.com/labstack/echo/v4"
)

func PostToggleManualMode(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	tx := bot.BeginNamed("Manual Mode")
	c.Set("manual-mode", tx)
	c.Get("manuel-mode")
	return nil
}

func AddAccountHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	number := c.Param("number")
	lang := c.Param("lang")

	lobby := "lobby"
	if bot.IsPioneers() {
		lobby = "lobby-pioneers"
	}
	accountGroup := fmt.Sprintf("%s_%s", lang, number)
	accountRes, err := gameforge.AddAccount(bot.GetDevice(), context.Background(), lobby, accountGroup, bot.GetBearerToken())
	if err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(500, "bearer-token:"+bot.GetBearerToken()+" Lobby: "+lobby+" Group: "+accountGroup+" "+err.Error()))
	}
	return c.JSON(http.StatusBadRequest, wrapper.SuccessResp(accountRes))
}

func GetAccountsHandler(c echo.Context) error {
	bot := c.Get("bot").(*wrapper.OGame)
	lobby := "lobby"
	if bot.IsPioneers() {
		lobby = "lobby-pioneers"
	}
	accounts, err := gameforge.GetUserAccounts(bot.GetClient(), context.Background(), lobby, bot.GetBearerToken())
	if err != nil {
		return c.JSON(http.StatusBadRequest, wrapper.ErrorResp(500, err.Error()))
	}
	return c.JSON(http.StatusBadRequest, wrapper.SuccessResp(accounts))
}
