package api

import (
	"encoding/xml"
	"fmt"
	"go-iptv/assets"
	"go-iptv/dto"
	"go-iptv/service"
	"go-iptv/until"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func MytvGetUserM3U8(c *gin.Context) {
	ts := c.Query("ts")             // 不存在时返回 ""
	deviceId := c.Query("deviceId") // 不存在时返回 ""

	if ts == "" || deviceId == "" {
		c.String(200, "参数错误1")
		return
	}

	clientIP := c.ClientIP()

	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "参数错误2")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)

	c.String(200, service.MytvGetUserM3U8(ts, deviceId, clientIP, host))
}

func MytvGetRssEpg(c *gin.Context) {
	deviceId := c.Param("deviceId")
	if deviceId == "" {
		c.Data(200, "text/xml", []byte(xml.Header+getQingh()))
		return
	}

	tv := service.MytvGetRssEpg(deviceId)

	output, err := xml.MarshalIndent(tv, "", "  ")
	if err != nil {
		log.Printf("生成XML失败: %v", err)
		c.Data(200, "text/xml", []byte(xml.Header+getQingh()))
		return
	}
	xmlData := []byte(xml.Header + string(output))

	c.Data(200, "text/xml", xmlData)
}

func MytvReleases(c *gin.Context) {
	c.JSON(200, service.MytvReleases())
}

func getQingh() string {
	res := dto.XmlTV{
		GeneratorName: "清和IPTV管理系统",
		GeneratorURL:  "https://www.qingh.xyz",
	}
	output, _ := xml.MarshalIndent(res, "", "  ")
	return string(output)
}

func BaseApk(c *gin.Context) {
	c.FileFromFS("baseApk/MyTV.apk", http.FS(assets.MyTVApk))
}

func BaseVersion(c *gin.Context) {
	c.String(200, string(assets.MyTVApkVersion))
}
