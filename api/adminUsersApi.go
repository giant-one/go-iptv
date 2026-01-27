package api

import (
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"go-iptv/until"
	"strconv"
	"time"

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
		case "submitextend":
			ids := c.PostFormArray("ids[]")
			if len(ids) == 0 {
				c.JSON(200, gin.H{"code": 0, "msg": "请选择要延期的用户账号", "type": "danger"})
				return
			}
			daysStr := c.PostForm("extend_days")
			if daysStr == "" {
				c.JSON(200, gin.H{"code": 0, "msg": "请输入延期天数", "type": "danger"})
				return
			}
			days, err := strconv.Atoi(daysStr)
			if err != nil || days <= 0 {
				c.JSON(200, gin.H{"code": 0, "msg": "延期天数必须为正整数", "type": "danger"})
				return
			}

			// 获取当前用户信息，计算新的过期时间
			var users []models.IptvUser
			dao.DB.Where("name in (?)", ids).Find(&users)

			// 批量更新过期时间
			for _, user := range users {
				var newExp int64
				if user.Exp == 0 {
					// 如果原来是永久的，从现在开始计算
					newExp = time.Now().Unix() + int64(days*86400)
				} else {
					// 如果原来有期限，在原基础上增加天数
					newExp = user.Exp + int64(days*86400)
				}

				dao.DB.Model(&models.IptvUser{}).Where("name = ?", user.Name).Updates(map[string]interface{}{
					"exp":    newExp,
					"status": 1,
				})
			}

			c.JSON(200, gin.H{"code": 1, "msg": "已延期选中的用户账号", "type": "success"})
			return
		}
	}
}
