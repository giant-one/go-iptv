package api

import (
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"go-iptv/until"
	"strconv"

	"github.com/gin-gonic/gin"
)

func EditUsers(c *gin.Context) {
	_, ok := until.GetAuthName(c)
	if !ok {
		c.JSON(200, dto.NewAdminRedirectDto())
		return
	}
	c.Request.ParseForm()
	params := c.Request.PostForm
	for k := range params {
		switch k {
		case "submitdel":
			ids := c.PostFormArray("ids[]")
			if len(ids) == 0 {
				c.JSON(200, gin.H{"code": 0, "msg": "请选择要删除的用户账号", "type": "danger"})
				return
			}
			dao.DB.Where("name in (?)", ids).Delete(&models.IptvUser{})
			c.JSON(200, gin.H{"code": 1, "msg": "已删除选中的用户账号", "type": "success"})
			return
		case "submitmodifymarks":
			ids := c.PostFormArray("ids[]")
			if len(ids) == 0 {
				c.JSON(200, gin.H{"code": 0, "msg": "请选择要修改备注的用户账号", "type": "danger"})
				return
			}
			marks := c.PostForm("marks")
			dao.DB.Model(&models.IptvUser{}).Where("name in (?)", ids).Update("marks", marks)
			c.JSON(200, gin.H{"code": 1, "msg": "已修改选中的用户账号的备注", "type": "success"})
			return
		case "submitforbidden":
			ids := c.PostFormArray("ids[]")
			if len(ids) == 0 {
				c.JSON(200, gin.H{"code": 0, "msg": "请选择要取消授权的用户账号", "type": "danger"})
				return
			}
			dao.DB.Model(&models.IptvUser{}).Where("name in (?)", ids).Updates(map[string]interface{}{
				"status": -1,
			})
			c.JSON(200, gin.H{"code": 1, "msg": "已取消选中的用户账号的授权", "type": "success"})
			return
		case "e_meals":
			ids := c.PostFormArray("ids[]")
			if len(ids) == 0 {
				c.JSON(200, gin.H{"code": 0, "msg": "请选择要修改套餐的用户账号", "type": "danger"})
				return
			}
			mealStr := c.PostForm("s_meals")
			mealID, err := strconv.Atoi(mealStr)
			if err != nil {
				c.JSON(200, gin.H{"code": 0, "msg": "套餐格式不正确", "type": "danger"})
				return
			}
			var meal models.IptvMeals
			err = dao.DB.Where("id = ?", mealID).First(&meal).Error
			if err != nil {
				c.JSON(200, gin.H{"code": 0, "msg": "选择的套餐不存在", "type": "danger"})
				return
			}

			dao.DB.Model(&models.IptvUser{}).Where("name in (?)", ids).Updates(map[string]interface{}{
				"meal":   meal.ID,
				"status": 999,
			})
			c.JSON(200, gin.H{"code": 1, "msg": "已修改选中的用户账号的套餐", "type": "success"})
			return
		}
	}
}
