package response

// 数据爬取规则
type Rule struct {
	RuleType string `json:"type"`
	Path     string
	Parser   string
	Attr     string
}
