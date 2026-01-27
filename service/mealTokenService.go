package service

import (
	"crypto/rand"
	"encoding/hex"
	"go-iptv/dao"
	"go-iptv/dto"
	"go-iptv/models"
	"go-iptv/until"
	"log"
	"net/url"
	"strconv"
	"time"
)

// GetMealTokens 获取套餐的所有token
func GetMealTokens(mealId int64) ([]dto.MealTokenDto, error) {
	var tokens []models.IptvMealToken
	if err := dao.DB.Where("meal_id = ?", mealId).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}

	var result []dto.MealTokenDto
	for _, token := range tokens {
		result = append(result, dto.MealTokenDto{
			ID:        token.ID,
			MealID:    token.MealID,
			Token:     token.Token,
			CreatedAt: token.CreatedAt,
			ExpiresAt: token.ExpiresAt,
			Status:    token.Status,
			Remark:    token.Remark,
		})
	}
	return result, nil
}

// CreateMealToken 创建新的token
func CreateMealToken(mealId int64, remark string, expireDays int64) (dto.MealTokenDto, error) {
	// 生成随机值
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return dto.MealTokenDto{}, err
	}
	randomStr := hex.EncodeToString(randomBytes)

	// 生成新的token
	aesData := AesData{
		I: mealId,
		R: randomStr,
	}
	aesDataStr, err := getAesdata(aesData)
	if err != nil {
		return dto.MealTokenDto{}, err
	}

	// 使用ChaCha20加密生成token
	aes := until.NewChaCha20(string(until.RssKey))
	token, err := aes.Encrypt(aesDataStr)
	if err != nil {
		return dto.MealTokenDto{}, err
	}
	log.Printf("[DEBUG]CreateMealToken - mealId=%d, RssKey=%s, aesDataStr=%s, token=%s",
		mealId, string(until.RssKey), aesDataStr, token)

	// 计算过期时间
	var expiresAt int64 = 0
	if expireDays > 0 {
		expiresAt = time.Now().Unix() + expireDays*24*60*60
	}

	// 创建token记录
	mealToken := models.IptvMealToken{
		MealID:     mealId,
		Token:      token,
		CreatedAt:  time.Now().Unix(),
		ExpiresAt:  expiresAt,
		Status:     1,
		Remark:     remark,
		ExpireDays: expireDays,
	}

	if err := dao.DB.Create(&mealToken).Error; err != nil {
		return dto.MealTokenDto{}, err
	}

	return dto.MealTokenDto{
		ID:         mealToken.ID,
		MealID:     mealToken.MealID,
		Token:      mealToken.Token,
		CreatedAt:  mealToken.CreatedAt,
		ExpiresAt:  mealToken.ExpiresAt,
		Status:     mealToken.Status,
		Remark:     mealToken.Remark,
		ExpireDays: mealToken.ExpireDays,
	}, nil
}

// UpdateMealToken 更新token信息
func UpdateMealToken(params url.Values) dto.ReturnJsonDto {
	tokenId := params.Get("token_id")
	if tokenId == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "Token ID不能为空", Type: "danger"}
	}

	tokenIdInt64, err := strconv.ParseInt(tokenId, 10, 64)
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "Token ID格式错误", Type: "danger"}
	}

	// 获取状态和备注
	status := params.Get("status")
	remark := params.Get("remark")

	// 更新token
	updates := make(map[string]interface{})
	if status != "" {
		statusInt64, err := strconv.ParseInt(status, 10, 64)
		if err != nil {
			return dto.ReturnJsonDto{Code: 0, Msg: "状态格式错误", Type: "danger"}
		}
		updates["status"] = statusInt64
	}
	if remark != "" {
		updates["remark"] = remark
	}

	if len(updates) > 0 {
		if err := dao.DB.Model(&models.IptvMealToken{}).Where("id = ?", tokenIdInt64).Updates(updates).Error; err != nil {
			return dto.ReturnJsonDto{Code: 0, Msg: "更新失败: " + err.Error(), Type: "danger"}
		}
	}

	return dto.ReturnJsonDto{Code: 1, Msg: "更新成功", Type: "success"}
}

// DeleteMealToken 删除token
func DeleteMealToken(tokenId int64) error {
	return dao.DB.Where("id = ?", tokenId).Delete(&models.IptvMealToken{}).Error
}

// DeleteMealTokenAPI 删除token API接口
func DeleteMealTokenAPI(params url.Values) dto.ReturnJsonDto {
	tokenId := params.Get("token_id")
	if tokenId == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "Token ID不能为空", Type: "danger"}
	}

	tokenIdInt64, err := strconv.ParseInt(tokenId, 10, 64)
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "Token ID格式错误", Type: "danger"}
	}

	if err := DeleteMealToken(tokenIdInt64); err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "删除失败: " + err.Error(), Type: "danger"}
	}

	return dto.ReturnJsonDto{Code: 1, Msg: "删除成功", Type: "success"}
}

// ExtendToken 延期token
func ExtendToken(tokenId int64, extendDays int64) error {
	// 获取当前token信息
	var token models.IptvMealToken
	if err := dao.DB.Where("id = ?", tokenId).First(&token).Error; err != nil {
		return err
	}

	// 计算新的过期时间
	var newExpiresAt int64
	if token.ExpiresAt == 0 {
		// 如果原来没有设置过期时间，则从现在开始计算
		newExpiresAt = time.Now().Unix() + extendDays*24*60*60
	} else {
		// 如果原来设置了过期时间，则在原来的基础上延长
		newExpiresAt = token.ExpiresAt + extendDays*24*60*60
	}

	// 更新过期时间和有效期天数
	updates := map[string]interface{}{
		"expires_at":  newExpiresAt,
		"expire_days": token.ExpireDays + extendDays,
	}

	return dao.DB.Model(&models.IptvMealToken{}).Where("id = ?", tokenId).Updates(updates).Error
}

// ExtendTokenAPI 延期token API接口
func ExtendTokenAPI(params url.Values) dto.ReturnJsonDto {
	tokenId := params.Get("token_id")
	if tokenId == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "Token ID不能为空", Type: "danger"}
	}

	tokenIdInt64, err := strconv.ParseInt(tokenId, 10, 64)
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "Token ID格式错误", Type: "danger"}
	}

	extendDaysStr := params.Get("extend_days")
	if extendDaysStr == "" {
		return dto.ReturnJsonDto{Code: 0, Msg: "延期天数不能为空", Type: "danger"}
	}

	extendDays, err := strconv.ParseInt(extendDaysStr, 10, 64)
	if err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "延期天数格式错误", Type: "danger"}
	}

	if extendDays <= 0 {
		return dto.ReturnJsonDto{Code: 0, Msg: "延期天数必须大于0", Type: "danger"}
	}

	if err := ExtendToken(tokenIdInt64, extendDays); err != nil {
		return dto.ReturnJsonDto{Code: 0, Msg: "延期失败: " + err.Error(), Type: "danger"}
	}

	return dto.ReturnJsonDto{Code: 1, Msg: "延期成功", Type: "success"}
}