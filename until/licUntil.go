package until

import (
	"encoding/json"
	"go-iptv/dao"
	"log"
	"os/exec"
	"strings"
)

func IsRunning() bool {
	cmd := exec.Command("bash", "-c", "ps -ef | grep '/license' | grep -v grep")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("检查 授权服务 进程出错: %v", err)
		return false
	}
	return strings.Contains(string(output), "license")
}

func RestartLic() bool {
	log.Println("♻️ 正在重启  授权服务  服务...")

	r := GetUrlData("http://127.0.0.1:82/licRestart")
	if strings.TrimSpace(r) == "" {
		log.Println("升级服务未启动")
		return false
	}
	if strings.TrimSpace(r) != "OK" {
		return false
	}

	ws, err := dao.ConLicense("ws://127.0.0.1:81/ws")
	if err != nil {
		log.Println("授权服务连接失败")
		return false
	}
	dao.WS = ws
	res, err := dao.WS.SendWS(dao.Request{Action: "getlic"})
	if err == nil {
		if err := json.Unmarshal(res.Data, &dao.Lic); err == nil {
			log.Println("license初始化成功")
			log.Println("机器码:", dao.Lic.ID)
		} else {
			log.Println("license信息解析错误:", err)
		}
	} else {
		log.Println("license初始化错误")
		return false
	}

	log.Println("✅  授权服务  已成功重启并重新连接")
	return true
}
