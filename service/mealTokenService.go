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
func CreateMealToken(mealId int64, remark string) (dto.MealTokenDto, error) {
	// 生成新的token
	aesData := AesData{
		I: mealId,
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

	// 创建token记录
	mealToken := models.IptvMealToken{
		MealID:    mealId,
		Token:     token,
		CreatedAt: time.Now().Unix(),
		Status:    1,
		Remark:    remark,
	}

	if err := dao.DB.Create(&mealToken).Error; err != nil {
		return dto.MealTokenDto{}, err
	}

	return dto.MealTokenDto{
		ID:        mealToken.ID,
		MealID:    mealToken.MealID,
		Token:     mealToken.Token,
		CreatedAt: mealToken.CreatedAt,
		ExpiresAt: mealToken.ExpiresAt,
		Status:    mealToken.Status,
		Remark:    mealToken.Remark,
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