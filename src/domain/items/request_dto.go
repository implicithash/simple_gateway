package items

type Request struct {
	Items []RequestItem `json:"requests"`
}

type RequestItem struct {
	Url    string  `json:"url"`
	Params Param `json:"params"`
}

type Param struct {
}
