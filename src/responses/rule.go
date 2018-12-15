package responses

// 数据爬取规则
type Rule struct {
	RuleType string
	Path     string
	Parser   string
	Attr     string
}
