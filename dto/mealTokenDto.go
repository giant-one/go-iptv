package dto

type MealTokenDto struct {
	ID        int64  `json:"id"`
	MealID    int64  `json:"meal_id"`
	Token     string `json:"token"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
	Status    int64  `json:"status"`
	Remark    string `json:"remark"`
}

type MealTokenListDto struct {
	Tokens []MealTokenDto `json:"tokens"`
}

type MealTokenCreateDto struct {
	MealID int64  `json:"meal_id"`
	Remark string `json:"remark"`
}

type MealTokenUpdateDto struct {
	ID     int64  `json:"id"`
	Status int64  `json:"status"`
	Remark string `json:"remark"`
}