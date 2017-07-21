package model

//Related stores the Title and URI for any related data (eg related publications on a dataset page)
type Related struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
}
