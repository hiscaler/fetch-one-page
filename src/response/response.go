package response

type Error struct {
	Message string `json:"message"`
}

type BaseResponse struct {
	Success bool `json:"success"`
}

// {"success": false, "error": {"message": "error message"}}
type FailResponse struct {
	Success bool  `json:"success"`
	Error   Error `json:"error"`
}

// {"success": true, "data": {"items": []}}
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    SuccessData `json:"data"`
}

type SuccessData struct {
	Items []Url `json:"items"`
}
