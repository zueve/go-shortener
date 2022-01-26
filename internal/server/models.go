package server

type Redirect struct {
	URL string `json:"url"`
}

type ResultString struct {
	Result string `json:"result"`
}


type URLRow struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
