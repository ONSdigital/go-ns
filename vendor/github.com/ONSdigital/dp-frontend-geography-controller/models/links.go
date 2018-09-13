package models

// Link contains the id and a link to a resource
type Link struct {
	ID   string `json:"id,omitempty"     bson:"id"`
	Href string `json:"href"             bson:"href"`
}
