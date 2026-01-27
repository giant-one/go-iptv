package bootstrap

import (
	"go-iptv/dao"
	"go-iptv/until"
	"log"
	"time"
)

func InitJwtKey() {
	// 读取配置文件
	hostname, _ := until.GetContainerID()
	until.JwtKey = []byte(until.Md5(hostname + time.Now().Format("2006-01-02 15:04:05")))
	cfg := dao.GetConfig()
	if cfg.Rss.Key == "" {
		cfg.Rss.Key = until.Md5(time.Now().Format("2006-01-02 15:04:05"))
		until.RssKey = []byte(cfg.Rss.Key)
		dao.SetConfig(cfg)
	} else {
		until.RssKey = []byte(cfg.Rss.Key)
	}
	log.Printf("[DEBUG]InitJwtKey - RssKey=%s", string(until.RssKey))
	// until.RssKey = []byte(until.Md5(hostname))
}
