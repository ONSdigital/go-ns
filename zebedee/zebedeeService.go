package zebedee

import (
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/zebedee/model"
)

// Service defines interface of zebedee service.
type Service interface {
	GetData(url string, requestContentID string) (data []byte, pageType string, err *common.ONSError)
	GetTaxonomy(url string, depth int, requestContentID string) ([]model.ContentNode, *common.ONSError)
	GetParents(url string, requestContentID string) ([]model.ContentNode, *common.ONSError)
	GetTimeSeries(url string, requestContentID string) (*model.TimeseriesPage, *common.ONSError)
}
