package service

import (
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"go-iptv/until"
	"log"
	"strconv"
	"time"
)

func MytvGetUserM3U8(ts, deviceId, clientIP, host string) string {
	var user models.IptvUser
	user.IP = clientIP
	user.DeviceID = deviceId
	user.Region = until.GetIpRegion(user.IP)

	if lastTime, err := strconv.ParseInt(ts, 10, 64); err == nil {
		user.LastTime = lastTime
	} else {
		user.LastTime = time.Now().Unix()
	}
	user = SaveUser(user)
	keySeed := ts + deviceId

	data, err := until.AESEncrypt(until.MytvM3u8(int64(user.Meal), host), keySeed)
	if err != nil {
		log.Println("mytv订阅加密失败: ", err)
	}
	return data
}

func MytvGetRssEpg(deviceid string) dto.XmlTV {
	var dbUser models.IptvUser
	res := dao.DB.Where("deviceid = ?", deviceid).First(&dbUser)
	if res.RowsAffected == 0 {
		return dto.XmlTV{
			GeneratorName: "清和IPTV管理系统",
			GeneratorURL:  "https://www.qingh.xyz",
		}
	}
	return until.GetEpg(dbUser.Meal)
}

func SaveUser(user models.IptvUser) models.IptvUser {
	var dbUser models.IptvUser
	res := dao.DB.Where("deviceid = ?", user.DeviceID).First(&dbUser)
	if res.RowsAffected == 0 {
		user.Name = int64(genName())
		var cfg = dao.GetConfig()
		switch cfg.App.NeedAuthor {
		case 0:
			user.Status = 999
			user.Marks = "自动授权"
		case 1:
			user.Status = -1
			user.Marks = "未授权"
		}
		user.Meal = 1000
		dao.DB.Model(&models.IptvUser{}).Create(&user)
		return user
	}

	dbUser.IP = user.IP
	dbUser.Region = user.Region
	dbUser.LastTime = user.LastTime

	dao.DB.Model(&models.IptvUser{}).Where("deviceid = ?", user.DeviceID).Updates(dbUser)
	return dbUser
}
