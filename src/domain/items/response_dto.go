package items

type Response struct {
	Items []*ResponseItem
}

type ResponseItem struct {
	Data string `json:"data"`
}