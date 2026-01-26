package service

import (
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"go-iptv/until"
	"net/url"
	"strconv"
	"time"
)

func SubmitAuthorForever(params url.Values, username string) dto.ReturnJsonDto {
	ids := params["ids[]"]
	meal := params.Get("meal")

	if len(ids) == 0 {
		return dto.ReturnJsonDto{Code: 0, Msg: "请选择用户", Type: "danger"}
	}
	if meal == "" || meal == "0" {
		return dto.ReturnJsonDto{Code: 0, Msg: "请选择套餐", Type: "danger"}
	}
	if !until.IsSafe(meal) {
		return dto.ReturnJsonDto{Code: 0, Msg: "输入不合法", Type: "danger"}
	}

	dao.DB.Model(&models.IptvUser{}).Where("name IN (?)", ids).Updates(map[string]interface{}{
		"meal":       meal,
		"status":     999,
		"exp":        0,
		"author":     username,
		"authortime": time.Now().Unix(),
		"marks":      username + "已授权",
	})
	return dto.ReturnJsonDto{
		Code: 1,
		Msg:  "操作成功",
		Type: "success",
	}
}

func SubmitAuthorWithDays(params url.Values, username string) dto.ReturnJsonDto {
	ids := params["ids[]"]
	meal := params.Get("meal")
	days := params.Get("author_days")

	if len(ids) == 0 {
		return dto.ReturnJsonDto{Code: 0, Msg: "请选择用户", Type: "danger"}
	}
	if meal == "" || meal == "0" {
		return dto.ReturnJsonDto{Code: 0, Msg: "请选择套餐", Type: "danger"}
	}
	if !until.IsSafe(meal) {
		return dto.ReturnJsonDto{Code: 0, Msg: "输入不合法", Type: "danger"}
	}
	if days == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "请输入授权天数", Type: "danger"}
	}
	if !until.IsSafe(days) {
		return dto.ReturnJsonDto{Code: 0, Msg: "授权天数输入不合法", Type: "danger"}
	}

	// 将天数转换为过期时间戳
	daysInt, err := strconv.Atoi(days)
	if err != nil || daysInt <= 0 {
		return dto.ReturnJsonDto{Code: 0, Msg: "授权天数必须为正整数", Type: "danger"}
	}

	// 计算过期时间戳（当前时间 + 天数）
	expTime := time.Now().AddDate(0, 0, daysInt).Unix()

	dao.DB.Model(&models.IptvUser{}).Where("name IN (?)", ids).Updates(map[string]interface{}{
		"meal":       meal,
		"status":     999,
		"exp":        expTime,
		"author":     username,
		"authortime": time.Now().Unix(),
		"marks":      username + "已授权(" + days + "天)",
	})
	return dto.ReturnJsonDto{
		Code: 1,
		Msg:  "操作成功",
		Type: "success",
	}
}

func ForbiddenUser(params url.Values) dto.ReturnJsonDto {
	ids := params["ids[]"]

	if len(ids) == 0 {
		return dto.ReturnJsonDto{Code: 0, Msg: "请选择用户", Type: "danger"}
	}
	dao.DB.Model(&models.IptvUser{}).Where("name IN (?)", ids).Updates(map[string]interface{}{
		"status": 0,
	})
	return dto.ReturnJsonDto{
		Code: 1,
		Msg:  "操作成功",
		Type: "success",
	}
}

func DelUsers(params url.Values) dto.ReturnJsonDto {
	ids := params["ids[]"]

	if len(ids) == 0 {
		return dto.ReturnJsonDto{Code: 0, Msg: "请选择用户", Type: "danger"}
	}
	dao.DB.Model(&models.IptvUser{}).
		Where("name IN ?", ids).
		Delete(&models.IptvUser{})

	return dto.ReturnJsonDto{
		Code: 1,
		Msg:  "操作成功",
		Type: "success",
	}
}

func DelUnAuthorOneDayBefore() dto.ReturnJsonDto {

	dao.DB.Model(&models.IptvUser{}).Where("status = ? and lasttime < ?", -1, time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location()).Unix()).Delete(&models.IptvUser{})

	return dto.ReturnJsonDto{
		Code: 1,
		Msg:  "操作成功",
		Type: "success",
	}
}

func DelAllUsers() dto.ReturnJsonDto {
	dao.DB.Model(&models.IptvUser{}).Where("status = ?", -1).Delete(&models.IptvUser{})
	return dto.ReturnJsonDto{
		Code: 1,
		Msg:  "操作成功",
		Type: "success",
	}
}
