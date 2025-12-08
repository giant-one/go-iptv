package bootstrap

import (
	"encoding/json"
	"go-iptv/dao"
	"log"
)

func InitLicense() {
	// dao.StartLicense()
	log.Println("引擎初始化中")
	err := dao.WS.Start("ws://127.0.0.1:81/ws")
	if err != nil {
		log.Println("引擎初始化错误: ", err)
		return
	}
	res, err := dao.WS.SendWS(dao.Request{Action: "getlic"})
	if err == nil {
		if err := json.Unmarshal(res.Data, &dao.Lic); err == nil {
			log.Println("引擎初始化成功")
			log.Println("机器码:", dao.Lic.ID)
		} else {
			log.Println("授权信息解析错误:", err)
		}
	} else {
		log.Println("引擎初始化错误")
		return
	}
}
