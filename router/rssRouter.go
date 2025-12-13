package router

import (
	"go-iptv/api"

	"github.com/gin-gonic/gin"
)

func RssRouter(r *gin.Engine, path string) {
	router := r.Group(path)
	{
		router.GET("/getRss/:token/paylist.m3u", api.GetRssM3u)
		router.GET("/getRss/:token/paylist.txt", api.GetRssTxt)
		router.GET("/ku9/:token/paylist.txt", api.GetRssTxtKu9)
		router.GET("/epg/:token/e.xml", api.GetRssEpg)

		router.GET("/r/:key/p.m3u", api.GetRssM3uShortURL)
		router.GET("/r/:key/p.txt", api.GetRssTxtShortURL)
		router.GET("/k/:key/p.txt", api.GetRssTxtKu9ShortURL)
		router.GET("/r/:key/e.xml", api.GetRssEpgShortURL)
	}
}
