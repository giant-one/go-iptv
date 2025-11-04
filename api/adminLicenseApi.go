package api

import (
	"go-iptv/dto"
	"go-iptv/service"
	"go-iptv/until"

	"github.com/gin-gonic/gin"
)

func License(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}
	c.Request.ParseForm()
	params := c.Request.PostForm
	var res dto.ReturnJsonDto

	for k := range params {
		switch k {
		case "importLic":
			res = service.ImportLicense(params)
		case "proxy":
			res = service.Proxy(params)
		case "resEng":
			res = service.ResEng()
		case "autoRes":
			res = service.AutoRes(params)
		case "disCh":
			res = service.DisCh(params)
		}

	}

	c.JSON(200, res)
}
