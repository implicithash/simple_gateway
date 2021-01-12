package items

// Response is a request response contains data of other GET responses
type Response struct {
	Items []*ResponseItem
}

// ResponseItem is an item of one request
type ResponseItem struct {
	Data string `json:"data"`
}
