package models

type IptvUser struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        int64  `gorm:"column:name" json:"name"`
	Mac         string `gorm:"column:mac" json:"mac"`
	DeviceID    string `gorm:"column:deviceid" json:"deviceid"`
	Model       string `gorm:"column:model" json:"model"`
	IP          string `gorm:"column:ip" json:"ip"`
	Region      string `gorm:"column:region" json:"region"`
	Exp         int64  `gorm:"column:exp" json:"exp"`
	VPN         int    `gorm:"column:vpn" json:"vpn"`
	IDChange    int    `gorm:"column:idchange" json:"idchange"`
	Author      string `gorm:"column:author" json:"author"`
	AuthorTime  int64  `gorm:"column:authortime" json:"authortime"`
	Status      int64  `gorm:"default:-1;column:status" json:"status"`
	LastTime    int64  `gorm:"column:lasttime" json:"lasttime"`
	Marks       string `gorm:"column:marks" json:"marks"`
	Meal        int64  `gorm:"column:meal" json:"meal"`
	LastTimeStr string `gorm:"-" json:"lasttime_str"`
	ExpDesc     string `gorm:"-" json:"expdesc"`    // 剩余xx天
	ExpDays     string `gorm:"-" json:"expdays"`    // 剩余天数
	StatusDesc  string `gorm:"-" json:"statusdesc"` // 状态描述
	NetType     string `gorm:"-" json:"nettype"`    // 网络类型
}

type IptvUserShow struct {
	ID          int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        int64  `gorm:"column:name" json:"name"`
	Mac         string `gorm:"column:mac" json:"mac"`
	DeviceID    string `gorm:"column:deviceid" json:"deviceid"`
	Model       string `gorm:"column:model" json:"model"`
	IP          string `gorm:"column:ip" json:"ip"`
	Region      string `gorm:"column:region" json:"region"`
	Exp         int64  `gorm:"column:exp" json:"exp"`
	VPN         int    `gorm:"column:vpn" json:"vpn"`
	IDChange    int    `gorm:"column:idchange" json:"idchange"`
	Author      string `gorm:"column:author" json:"author"`
	AuthorTime  int64  `gorm:"column:authortime" json:"authortime"`
	Status      int64  `gorm:"default:-1;column:status" json:"status"`
	LastTime    int64  `gorm:"column:lasttime" json:"lasttime"`
	Marks       string `gorm:"column:marks" json:"marks"`
	Meal        int64  `gorm:"column:meal" json:"meal"`
	MealName    string `gorm:"->;column:mealname" json:"mealname"`
	LastTimeStr string `gorm:"-" json:"lasttime_str"`
	ExpDesc     string `gorm:"-" json:"expdesc"`    // 剩余xx天
	ExpDays     string `gorm:"-" json:"expdays"`    // 剩余天数
	StatusDesc  string `gorm:"-" json:"statusdesc"` // 状态描述
	NetType     string `gorm:"-" json:"nettype"`    // 网络类型
}

// 添加联合唯一约束
func (IptvUser) TableName() string {
	return "iptv_users"
}

func (IptvUserShow) TableName() string {
	return "iptv_users"
}
