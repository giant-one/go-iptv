package api

import (
	"encoding/xml"
	"fmt"
	"go-iptv/dto"
	"go-iptv/service"
	"go-iptv/until"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetRssUrl(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}
	id := c.PostForm("id")

	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)

	getnewkey, exists := c.GetPostForm("getnewkey")
	if exists && getnewkey != "" {
		c.JSON(200, service.GetRssUrl(getnewkey, host, true))
		return
	}

	c.JSON(200, service.GetRssUrl(id, host, false))
}

func GetRssM3u(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)

	c.String(200, service.GetRss(token, host, "m"))
}

func GetRssM3uShortURL(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.String(200, "key参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)
	token := service.GetRssToken(key)
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	c.String(200, service.GetRss(token, host, "m"))
}

func GetRssTxtShortURL(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.String(200, "key参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)
	token := service.GetRssToken(key)
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	c.String(200, service.GetRss(token, host, "t"))
}

func GetRssTxt(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)

	c.String(200, service.GetRss(token, host, "t"))
}

func GetRssTxtKu9ShortURL(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.String(200, "key参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)
	token := service.GetRssToken(key)
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	c.String(200, service.GetTxtKu9(token, host))
}

func GetRssTxtKu9(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)

	c.String(200, service.GetTxtKu9(token, host))
}

func GetRssEpgShortURL(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.String(200, "key参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)
	token := service.GetRssToken(key)
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	tv := service.GetRssEpg(token, host)

	output, err := xml.MarshalIndent(tv, "", "  ")
	if err != nil {
		c.String(200, "生成XML失败: %v", err)
		return
	}

	// 加上 XML 文件头
	xmlData := []byte(xml.Header + string(output))

	c.Data(200, "text/xml", xmlData)
}

// GetRssEpg 处理获取TXT格式RSS EPG的请求
// 参数:
//   - c: Gin框架的上下文对象，包含请求和响应信息
func GetRssEpg(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.String(200, "token 参数不存在")
		return
	}
	scheme := GetClientScheme(c)

	host := c.Request.Host
	if !until.IsValidHost(host) {
		c.String(200, "host不合法")
		return
	}
	host = fmt.Sprintf("%s://%s", scheme, host)

	tv := service.GetRssEpg(token, host)

	output, err := xml.MarshalIndent(tv, "", "  ")
	if err != nil {
		c.String(200, "生成XML失败: %v", err)
		return
	}

	// 加上 XML 文件头
	xmlData := []byte(xml.Header + string(output))

	c.Data(200, "text/xml", xmlData)
}

func GetClientScheme(c *gin.Context) string {
	// 1) X-Forwarded-Proto（可能是 "https" 或 "http"，也可能是逗号分隔的列表）
	if xf := c.Request.Header.Get("X-Forwarded-Proto"); xf != "" {
		// 取第一个值，移除空格，小写
		parts := strings.Split(xf, ",")
		if len(parts) > 0 {
			return strings.ToLower(strings.TrimSpace(parts[0]))
		}
	}
	if xf := c.Request.Header.Get("X-Forwarded-Scheme"); xf != "" {
		// 取第一个值，移除空格，小写
		parts := strings.Split(xf, ",")
		if len(parts) > 0 {
			return strings.ToLower(strings.TrimSpace(parts[0]))
		}
	}

	// 2) Forwarded: 表示形式如: Forwarded: for=192.0.2.60;proto=https;by=203.0.113.43
	if f := c.Request.Header.Get("Forwarded"); f != "" {
		// 简单查找 proto= 后面的值（更严格的解析可用正则或更完整解析）
		// 例如 "for=..., proto=https; ..." 或 ";proto=https"
		if i := strings.Index(strings.ToLower(f), "proto="); i != -1 {
			// 从 proto= 后面截取到下一个分号或逗号或结尾
			v := f[i+len("proto="):]
			end := len(v)
			for j, ch := range v {
				if ch == ';' || ch == ',' {
					end = j
					break
				}
			}
			return strings.ToLower(strings.TrimSpace(v[:end]))
		}
	}

	// 3) X-Forwarded-SSL: on 表示 https（一些旧代理会设置）
	if xfs := strings.ToLower(c.Request.Header.Get("X-Forwarded-SSL")); xfs == "on" {
		return "https"
	}

	// 4) 回退：检查当前连接是否使用 TLS（适用于没有代理或直连）
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}
