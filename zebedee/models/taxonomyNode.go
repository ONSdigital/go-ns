package models

// TaxonomyNode represents a taxonomy node
type TaxonomyNode struct {
	URI         string          `json:"uri"`
	Description NodeDescription `json:"description"`
	Type        string          `json:"type"`
}

// NodeDescription represents a node description
type NodeDescription struct {
	Title string `json:"title"`
}
