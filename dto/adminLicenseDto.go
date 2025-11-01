package dto

type AdminLicenseDto struct {
	LoginUser string `json:"loginuser"`
	Title     string `json:"title"`
	Proxy     int64  `json:"proxy"`
	Port      int64  `json:"port"`
	ProxyAddr string `json:"proxy_addr"`
	Lic       Lic    `json:"lic"`
	Status    int64  `json:"status"`
	Online    int64  `json:"online"`
	Version   string `json:"version"`
}

type Lic struct {
	ID     string `json:"id"`
	Tpye   int64  `json:"type"`
	Exp    int64  `json:"exp"`
	ExpStr string `json:"exp_str"`
}
