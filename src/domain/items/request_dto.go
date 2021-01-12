package items

// Request is a array of requests
type Request struct {
	Items []RequestItem `json:"requests"`
}

// RequestItem is an item of request array
type RequestItem struct {
	URL    string `json:"url"`
	Params Param  `json:"params"`
}

// Param is possible parameters of request
type Param struct {
}
