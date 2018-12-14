package response

type Prop struct {
	Name      string
	ValueType string
	Title     string
	Required  bool
	Rules     []Rule
}
