package chromedp

type ChromeDp struct {
	Headless   bool
	DisableGPU bool `json:"disable-gpu"`
}
