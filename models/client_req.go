package models

// /kill接口请求体
type KillRequest struct {
	JobId int64 `json:"jobId"`
}

// /log接口的请求体
type LogRequest struct {
	LogDataTim  int64 `json:"logDataTim"`
	LogId       int64 `json:"LogId"`
	FromLineNum int   `json:"fromLineNum"`
}

// Executor返回的日志包
type LogResultContent struct {
	FromLineNum int    `json:"fromLineNum"`
	ToLineNum   int    `json:"toLineNum"`
	IsEnd       bool   `json:"isEnd"`
	LogContent  string `json:"logContent"`
}
