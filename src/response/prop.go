package response

// 爬取属性定义
type Prop struct {
	Name      string
	ValueType string `json:"type"`
	Label     string
	Required  bool
	Rules     []Rule
	Pipe      string
}
