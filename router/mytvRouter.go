package router

import (
	"go-iptv/api"

	"github.com/gin-gonic/gin"
)

func MytvRouter(r *gin.Engine, path string) {
	router := r.Group(path)
	{
		router.GET("/m3u8", api.MytvGetUserM3U8)
		router.GET("/:deviceId/e.xml", api.MytvGetRssEpg)
		router.GET("/releases", api.MytvReleases)
		router.GET("/baseApk", api.BaseApk)
		router.GET("/baseVersion", api.BaseVersion)
	}
}
