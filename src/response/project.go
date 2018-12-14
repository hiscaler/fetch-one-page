package response

type Project struct {
	Id string
	Name string
	Title string
	PageRenderMethod string
	UseAgent bool
	Props []Prop
}
