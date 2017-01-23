package model

import (
	renderModel "github.com/ONSdigital/dp-frontend-models/model"
)

type ContentNode struct {
	URI         string          `json:"uri"`
	Description PageDescription `json:"description"`
	PageType    string          `json:"type"`
	Children    []ContentNode   `json:"children,omitempty"`
}

// Recursively convert a list zebedee content nodes to a list of renderer taxonomy nodes.
func (node *ContentNode) Map() renderModel.TaxonomyNode {
	var taxonomyNode renderModel.TaxonomyNode

	taxonomyNode.Title = node.Description.Title
	taxonomyNode.URI = node.URI

	if len(node.Children) > 0 {
		var children = make([]renderModel.TaxonomyNode, 0)
		for _, childNode := range node.Children {
			children = append(children, childNode.Map())
		}
		taxonomyNode.Children = children
	}
	return taxonomyNode
}
