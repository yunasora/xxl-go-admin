package models

// 全局通用返回结构
type XxlResponse struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Content interface{} `json:"content,omitempty"`
}

type RegistryParam struct {
	RegistryGroup string `json:"registryGroup" form:"registryGroup"`
	RegistryKey   string `json:"registryKey" form:"registryKey"`
	RegistryValue string `json:"registryValue" form:"registryValue"`
}

type CallbackRequest struct {
	LogId       int64  `json:"logId"`
	LogDateTime int64  `json:"logDateTime"`
	HandleCode  int    `json:"handleCode"` // 200 为成功
	HandleMsg   string `json:"handleMsg"`
}
