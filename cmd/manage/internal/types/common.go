package types

// 初始化接口
type InitResponse struct {
	AdminMenu []AdminMenu `json:"menu"`
	AdminNode []string    `json:"node"`

	CurrTime        string              `json:"curr_time"`
	CurrentTimezone string              `json:"curr_timezone"`
	InitTimezone    []map[string]string `json:"init_timezone"`

	IsSuperAdmin bool `json:"is_super"` // 是否超管

	PwdSwitch int `json:"pwd_switch"` // 授权码开关

}

// ####################    时区     ########################
type TimezoneResponse struct {
	CurrTime        string              `json:"curr_time"`
	CurrentTimezone string              `json:"curr_timezone"`
	InitTimezone    []map[string]string `json:"init_timezone"`
}

type TimezoneSetResponse struct {
	Time string `json:"time" `
}
