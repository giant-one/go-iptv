package service

import (
	"encoding/json"
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/until"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

func ImportLicense(params url.Values) dto.ReturnJsonDto {
	lickey := params.Get("lickey")
	if lickey == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "lickey不能为空", Type: "danger"}
	}

	oldLickey := until.ReadFile("/config/license.lic")

	err := os.WriteFile("/config/license.lic", []byte(lickey), 0644)
	if err != nil {
		os.WriteFile("/config/license.lic", []byte(oldLickey), 0644)
		return dto.ReturnJsonDto{Code: 0, Msg: "文件写入失败: " + err.Error(), Type: "danger"}
	}

	res, err := dao.WS.SendWS(dao.Request{Action: "reloadLic"})
	if err != nil {
		os.WriteFile("/config/license.lic", []byte(oldLickey), 0644)
		return dto.ReturnJsonDto{Code: 0, Msg: "授权失败: " + err.Error(), Type: "danger"}
	} else if res.Code == 1 {
		//授权成功
	} else if res.Code != 1 {
		os.WriteFile("/config/license.lic", []byte(oldLickey), 0644)
		return dto.ReturnJsonDto{Code: 0, Msg: "授权失败: " + res.Msg, Type: "danger"}
	}
	return dto.ReturnJsonDto{Code: 1, Msg: "授权成功", Type: "success"}
}

func Proxy(params url.Values) dto.ReturnJsonDto {
	cfg := dao.GetConfig()
	if dao.Lic.Tpye == 0 {
		cfg.Proxy.Status = 0

		dao.SetConfig(cfg)
		dao.WS.SendWS(dao.Request{Action: "stopProxy"})
		return dto.ReturnJsonDto{Code: 0, Msg: "未授权", Type: "danger"}
	}

	port := params.Get("port")
	proxy := params.Get("proxy")
	pAddr := params.Get("pAddr")

	if proxy == "1" || proxy == "true" || proxy == "on" {
		if port == "" {
			return dto.ReturnJsonDto{Code: 0, Msg: "端口不能为空", Type: "danger"}
		}

		if pAddr == "" {
			return dto.ReturnJsonDto{Code: 0, Msg: "地址不能为空", Type: "danger"}
		}

		portInt64, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			return dto.ReturnJsonDto{Code: 0, Msg: "port为数字", Type: "danger"}
		}
		if portInt64 < 82 || portInt64 > 65535 {
			return dto.ReturnJsonDto{Code: 0, Msg: "port为82-65535", Type: "danger"}
		}
		res, err := dao.WS.SendWS(dao.Request{Action: "startProxy"})
		if err != nil {
			cfg.Proxy.Status = 0
			cfg.Proxy.Port = portInt64
			cfg.Proxy.PAddr = pAddr

			dao.SetConfig(cfg)
			return dto.ReturnJsonDto{Code: 0, Msg: "启动失败: " + err.Error(), Type: "danger"}
		} else if res.Code == 1 {
			cfg.Proxy.Status = 1
			cfg.Proxy.Port = portInt64
			cfg.Proxy.PAddr = pAddr

			dao.SetConfig(cfg)
			go until.CleanAutoCacheAll() // 清理缓存
			return dto.ReturnJsonDto{Code: 1, Msg: "启动成功", Type: "success"}
		} else if res.Code != 1 {
			cfg.Proxy.Status = 0
			cfg.Proxy.Port = portInt64
			cfg.Proxy.PAddr = pAddr

			dao.SetConfig(cfg)
			return dto.ReturnJsonDto{Code: 0, Msg: "启动失败: " + res.Msg, Type: "danger"}
		} else {
			cfg.Proxy.Status = 0
			cfg.Proxy.Port = portInt64
			cfg.Proxy.PAddr = pAddr

			dao.SetConfig(cfg)
			dao.WS.SendWS(dao.Request{Action: "stopProxy"})
			go until.CleanAutoCacheAll()
			return dto.ReturnJsonDto{Code: 0, Msg: "启动失败: " + res.Msg, Type: "danger"}
		}
	} else {
		cfg.Proxy.Status = 0

		dao.SetConfig(cfg)
		dao.WS.SendWS(dao.Request{Action: "stopProxy"})
		go until.CleanAutoCacheAll()
		return dto.ReturnJsonDto{Code: 1, Msg: "停止成功", Type: "success"}
	}

}

func ResEng() dto.ReturnJsonDto {
	if dao.RestartLic() {
		return dto.ReturnJsonDto{Code: 1, Msg: "重启成功", Type: "success"}
	}
	return dto.ReturnJsonDto{Code: 0, Msg: "重启失败", Type: "danger"}
}

func AutoRes(params url.Values) dto.ReturnJsonDto {
	autoRes := params.Get("autoRes")
	cfg := dao.GetConfig()
	if dao.Lic.Tpye == 0 {
		cfg.Resolution.Auto = 0
		dao.SetConfig(cfg)
		return dto.ReturnJsonDto{Code: 0, Msg: "未授权", Type: "danger"}
	}
	if autoRes == "1" || autoRes == "true" || autoRes == "on" {
		cfg.Resolution.Auto = 1
	} else {
		cfg.Resolution.Auto = 0
	}
	dao.SetConfig(cfg)
	return dto.ReturnJsonDto{Code: 1, Msg: "设置成功", Type: "success"}
}

func DisCh(params url.Values) dto.ReturnJsonDto {
	disCh := params.Get("disCh")
	cfg := dao.GetConfig()
	if dao.Lic.Tpye == 0 {
		cfg.Resolution.DisCh = 0
		dao.SetConfig(cfg)
		return dto.ReturnJsonDto{Code: 0, Msg: "未授权", Type: "danger"}
	}
	if disCh == "1" || disCh == "true" || disCh == "on" {
		cfg.Resolution.DisCh = 1
	} else {
		cfg.Resolution.DisCh = 0
	}
	dao.SetConfig(cfg)
	return dto.ReturnJsonDto{Code: 1, Msg: "设置成功", Type: "success"}
}

func EpgFuzz(params url.Values) dto.ReturnJsonDto {
	epgFuzz := params.Get("epgFuzz")
	cfg := dao.GetConfig()
	if dao.Lic.Tpye == 0 {
		cfg.Epg.Fuzz = 0
		dao.SetConfig(cfg)
		return dto.ReturnJsonDto{Code: 0, Msg: "未授权", Type: "danger"}
	}
	if epgFuzz == "1" || epgFuzz == "true" || epgFuzz == "on" {
		cfg.Epg.Fuzz = 1
	} else {
		cfg.Epg.Fuzz = 0
	}
	dao.SetConfig(cfg)
	return dto.ReturnJsonDto{Code: 1, Msg: "设置成功", Type: "success"}
}

func Register(params url.Values) dto.ReturnJsonDto {
	name := params.Get("name")
	pwd := params.Get("pwd")
	pwd2 := params.Get("pwd2")

	if name == "" || pwd == "" || pwd2 == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "用户名或密码不能为空", Type: "danger"}
	}
	emailSimple := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailSimple.MatchString(name) {
		return dto.ReturnJsonDto{Code: 0, Msg: "邮箱格式不正确", Type: "danger"}
	}

	if pwd != pwd2 {
		return dto.ReturnJsonDto{Code: 0, Msg: "两次输入的密码不一致", Type: "danger"}
	}

	res, err := dao.WS.SendWS(dao.Request{Action: "register", Data: dto.LoginDto{
		Name: name,
		Pwd:  pwd,
		Pwd2: pwd2,
	}})
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "连接服务器失败", Type: "danger"}
	} else if res.Code != 1 {
		return dto.ReturnJsonDto{Code: 0, Msg: res.Msg, Type: "danger"}
	}

	return dto.ReturnJsonDto{Code: 1, Msg: res.Msg, Type: "success"}
}

func Login(params url.Values) dto.ReturnJsonDto {
	name := params.Get("name")
	pwd := params.Get("pwd")

	if name == "" || pwd == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "用户名或密码不能为空", Type: "danger"}
	}

	res, err := dao.WS.SendWS(dao.Request{Action: "login", Data: dto.LoginDto{
		Name: name,
		Pwd:  pwd,
	}})
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "连接服务器失败", Type: "danger"}
	} else if res.Code != 1 {
		return dto.ReturnJsonDto{Code: 0, Msg: res.Msg, Type: "danger"}
	} else {
		if err := json.Unmarshal(res.Data, &dao.Lic); err != nil {
			log.Println("⚠️ 无法解析服务器返回的key:", err)
			return dto.ReturnJsonDto{Code: 0, Msg: "连接服务器失败", Type: "danger"}
		}
	}

	return dto.ReturnJsonDto{Code: 1, Msg: "登录成功", Type: "success"}
}

func Logout() dto.ReturnJsonDto {
	res, err := dao.WS.SendWS(dao.Request{Action: "logout"})
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "连接服务器失败", Type: "danger"}
	} else if res.Code != 1 {
		return dto.ReturnJsonDto{Code: 0, Msg: res.Msg, Type: "danger"}
	} else {
		if err := json.Unmarshal(res.Data, &dao.Lic); err != nil {
			log.Println("⚠️ 无法解析服务器返回的key:", err)
			return dto.ReturnJsonDto{Code: 0, Msg: "连接服务器失败", Type: "danger"}
		}
	}
	return dto.ReturnJsonDto{Code: 1, Msg: "退出成功", Type: "success"}
}
