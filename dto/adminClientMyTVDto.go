package dto

type AdminClientMyTVDto struct {
	LoginUser   string `json:"loginuser"`
	Title       string `json:"title"`
	MyTV        MyTV   `json:"mytv"`
	IconUrl     string `json:"iconurl"`
	UpSize      string `json:"upsize"`
	BaseVerson  string `json:"baseverson"`
	ApkUrl      string `json:"apkurl"`
	ApkName     string `json:"apkname"`
	ServerUrl   string `json:"serverurl"`
	BuildStatus int64  `json:"buildstatus"` // APK编译状态
}
