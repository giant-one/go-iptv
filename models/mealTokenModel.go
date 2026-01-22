package models

type IptvMealToken struct {
	ID        int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MealID    int64  `gorm:"column:meal_id;not null" json:"meal_id"`
	Token     string `gorm:"column:token;not null;unique" json:"token"`
	CreatedAt int64  `gorm:"column:created_at;not null" json:"created_at"`
	ExpiresAt int64  `gorm:"column:expires_at" json:"expires_at"`
	Status    int64  `gorm:"column:status;not null;default:1" json:"status"`
	Remark    string `gorm:"column:remark" json:"remark"`
	ExpireDays int64 `gorm:"column:expire_days;default:0" json:"expire_days"`
}

func (IptvMealToken) TableName() string {
	return "iptv_meal_tokens"
}