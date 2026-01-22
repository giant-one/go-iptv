package api

import (
	"go-iptv/dto"
	"go-iptv/service"
	"go-iptv/until"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMealTokens 获取套餐的所有token
func GetMealTokens(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}

	mealId := c.Param("meal_id")
	if mealId == "" {
		c.JSON(200, dto.ReturnJsonDto{Code: 0, Msg: "套餐ID不能为空", Type: "danger"})
		return
	}

	mealIdInt64, err := strconv.ParseInt(mealId, 10, 64)
	if err != nil {
		c.JSON(200, dto.ReturnJsonDto{Code: 0, Msg: "套餐ID格式错误", Type: "danger"})
		return
	}

	tokens, err := service.GetMealTokens(mealIdInt64)
	if err != nil {
		c.JSON(200, dto.ReturnJsonDto{Code: 0, Msg: "获取token列表失败: " + err.Error(), Type: "danger"})
		return
	}

	c.JSON(200, dto.ReturnJsonDto{Code: 1, Data: dto.MealTokenListDto{Tokens: tokens}, Msg: "获取成功", Type: "success"})
}

// CreateMealToken 创建新的token
func CreateMealToken(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}

	mealId := c.Param("meal_id")
	if mealId == "" {
		c.JSON(200, dto.ReturnJsonDto{Code: 0, Msg: "套餐ID不能为空", Type: "danger"})
		return
	}

	mealIdInt64, err := strconv.ParseInt(mealId, 10, 64)
	if err != nil {
		c.JSON(200, dto.ReturnJsonDto{Code: 0, Msg: "套餐ID格式错误", Type: "danger"})
		return
	}

	remark := c.PostForm("remark")

	token, err := service.CreateMealToken(mealIdInt64, remark)
	if err != nil {
		c.JSON(200, dto.ReturnJsonDto{Code: 0, Msg: "创建token失败: " + err.Error(), Type: "danger"})
		return
	}

	c.JSON(200, dto.ReturnJsonDto{Code: 1, Data: token, Msg: "创建成功", Type: "success"})
}

// UpdateMealToken 更新token
func UpdateMealToken(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}

	c.Request.ParseForm()
	params := c.Request.PostForm

	res := service.UpdateMealToken(params)
	c.JSON(200, res)
}

// DeleteMealToken 删除token
func DeleteMealToken(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}

	c.Request.ParseForm()
	params := c.Request.PostForm

	res := service.DeleteMealTokenAPI(params)
	c.JSON(200, res)
}