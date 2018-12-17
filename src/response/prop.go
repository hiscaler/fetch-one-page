package response

// 爬取属性定义
type Prop struct {
	Name      string
	ValueType string
	Title     string
	Required  bool
	Rules     []Rule
}
