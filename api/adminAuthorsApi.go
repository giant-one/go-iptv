package api

import (
	"go-iptv/dto"
	"go-iptv/service"
	"go-iptv/until"

	"github.com/gin-gonic/gin"
)

func Authors(c *gin.Context) {
	username, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}
	c.Request.ParseForm()
	params := c.Request.PostForm
	var res dto.ReturnJsonDto

	for k := range params {
		switch k {
		case "submitauthorforever":
			res = service.SubmitAuthorForever(params, username)
		case "submitauthor":
			res = service.SubmitAuthorWithDays(params, username)
		case "submitforbidden":
			res = service.ForbiddenUser(params)
		case "submitdelonedaybefor":
			res = service.DelUnAuthorOneDayBefore()
		case "submitdel":
			res = service.DelUsers(params)
		case "submitdelall":
			res = service.DelAllUsers()
		}

	}
	c.JSON(200, res)
}
