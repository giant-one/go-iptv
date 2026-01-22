package service

import (
	"encoding/json"
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"go-iptv/until"
	"log"
	"time"
)

type RssUrl struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type AesData struct {
	I int64  `json:"i"`
	R string `json:"r,omitempty"`
}

func getAesdata(aesData AesData) (string, error) {
	jsonBytes, err := json.Marshal(aesData)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func getAesType(jsonStr string) (AesData, error) {
	var data AesData

	// 反序列化（字符串 -> 结构体）
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func GetRssUrl(id, host string, getnewkey bool) dto.ReturnJsonDto {
	var res []RssUrl

	var meal models.IptvMeals
	if err := dao.DB.Model(&models.IptvMeals{}).Where("id = ? and status = 1", id).First(&meal).Error; err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "未找到上线套餐", Type: "danger"}
	}

	aesData := AesData{
		I: meal.ID,
	}
	aesDataStr, err := getAesdata(aesData)
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "生成key失败", Type: "danger"}
	}
	cfg := dao.GetConfig()
	if getnewkey {
		cfg.Rss.Key = until.Md5(time.Now().Format("2006-01-02 15:04:05"))
		until.RssKey = []byte(cfg.Rss.Key)
		dao.SetConfig(cfg)
	}

	aes := until.NewChaCha20(string(until.RssKey))
	token, err := aes.Encrypt(aesDataStr)
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "生成链接失败" + err.Error(), Type: "danger"}
	}

	if cfg.System.ShortURL == 1 && dao.Lic.Type != 0 {
		wsRes, err := dao.WS.SendWS(dao.Request{Action: "getShortURLKey", Data: token})
		if err == nil && wsRes.Code == 1 {
			var key string
			if err := json.Unmarshal(wsRes.Data, &key); err != nil {
				log.Println("短订阅key解析错误:", err)
			} else {
				res = append(res, RssUrl{Type: "m3u8", Url: host + "/r/" + key + "/p.m3u"})
				res = append(res, RssUrl{Type: "txt", Url: host + "/r/" + key + "/p.txt"})
				res = append(res, RssUrl{Type: "ku9", Url: host + "/k/" + key + "/p.txt"})
				res = append(res, RssUrl{Type: "epg", Url: host + "/r/" + key + "/e.xml"})
				return dto.ReturnJsonDto{Code: 1, Msg: "订阅生成成功", Type: "success", Data: res}
			}
		}
	}

	res = append(res, RssUrl{Type: "m3u8", Url: host + "/getRss/" + token + "/paylist.m3u"})
	res = append(res, RssUrl{Type: "txt", Url: host + "/getRss/" + token + "/paylist.txt"})
	res = append(res, RssUrl{Type: "ku9", Url: host + "/ku9/" + token + "/paylist.txt"})
	res = append(res, RssUrl{Type: "epg", Url: host + "/epg/" + token + "/e.xml"})

	return dto.ReturnJsonDto{Code: 1, Msg: "订阅生成成功", Type: "success", Data: res}
}

func GetRssToken(key string) string {
	cfg := dao.GetConfig()
	if cfg.System.ShortURL == 1 && dao.Lic.Type != 0 {
		wsRes, err := dao.WS.SendWS(dao.Request{Action: "getShortURLToken", Data: key})
		if err == nil && wsRes.Code == 1 {
			var token string
			if err := json.Unmarshal(wsRes.Data, &token); err != nil {
				log.Println("短订阅token解析错误:", err)
				return ""
			} else {
				return token
			}
		}
	}
	return ""
}

func GetRss(token, host, t string) string {

	aes := until.NewChaCha20(string(until.RssKey))
	jsonStr, err := aes.Decrypt(token)
	if err != nil {
		return "订阅失败,token解密错误"
	}
	aesData, err := getAesType(jsonStr)
	if err != nil {
		return "订阅失败，token读取错误"
	}
	if t == "t" {
		return until.GetTxt(aesData.I)
	} else {
		return until.GetM3u8(aesData.I, host, token)
	}
}

func GetTxtKu9(token, host string) string {
	aes := until.NewChaCha20(string(until.RssKey))
	jsonStr, err := aes.Decrypt(token)
	if err != nil {
		return "订阅失败,token解密错误"
	}
	aesData, err := getAesType(jsonStr)
	if err != nil {
		return "订阅失败，token读取错误"
	}
	return until.GetTxtKu9(aesData.I)
}

func GetRssEpg(token, host string) dto.XmlTV {

	res := dto.XmlTV{
		GeneratorName: "清和IPTV管理系统",
		GeneratorURL:  "https://www.qingh.xyz",
	}
	aes := until.NewChaCha20(string(until.RssKey))
	jsonStr, err := aes.Decrypt(token)
	if err != nil {
		return res
	}
	aesData, err := getAesType(jsonStr)
	if err != nil {
		return res
	}
	return until.GetEpg(aesData.I)
}
