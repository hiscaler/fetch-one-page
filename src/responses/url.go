package responses

// 单条爬取地址
type Url struct {
	Id      string
	Url     string
	Status  int
	Project Project
}
