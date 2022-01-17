package server

type Redirect struct {
	URL string `json:"url,required"` // required doesn't work
}

type ResultString struct {
	Result string `json:"result"`
}
