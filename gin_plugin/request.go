package gin_plugin

// Response http 统一响应结构体
type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}
