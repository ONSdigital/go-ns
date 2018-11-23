package zebedeeMapper

import (
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageStatic"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/data"
)

// StaticDatasetLandingPage is a StaticDatasetLandingPage representation
type StaticDatasetLandingPage datasetLandingPageStatic.Page

// MapZebedeeDatasetLandingPageToFrontendModel maps a zebedee response struct into a frontend model to be used for rendering
func MapZebedeeDatasetLandingPageToFrontendModel(dlp data.DatasetLandingPage, bcs []data.Breadcrumb, ds []data.Dataset) StaticDatasetLandingPage {

	var sdlp StaticDatasetLandingPage

	sdlp.Type = dlp.Type
	sdlp.URI = dlp.URI
	sdlp.Metadata.Title = dlp.Description.Title
	sdlp.Metadata.Description = dlp.Description.Summary

	for _, d := range dlp.RelatedDatasets {
		sdlp.DatasetLandingPage.Related.Datasets = append(sdlp.DatasetLandingPage.Related.Datasets, model.Related(d))
	}

	for _, d := range dlp.RelatedDocuments {
		sdlp.DatasetLandingPage.Related.Publications = append(sdlp.DatasetLandingPage.Related.Publications, model.Related(d))
	}

	for _, d := range dlp.RelatedMethodology {
		sdlp.DatasetLandingPage.Related.Methodology = append(sdlp.DatasetLandingPage.Related.Methodology, model.Related(d))
	}
	for _, d := range dlp.RelatedMethodologyArticle {
		sdlp.DatasetLandingPage.Related.Methodology = append(sdlp.DatasetLandingPage.Related.Methodology, model.Related(d))
	}

	for _, d := range dlp.RelatedLinks {
		sdlp.DatasetLandingPage.Related.Links = append(sdlp.DatasetLandingPage.Related.Links, model.Related(d))
	}

	sdlp.DatasetLandingPage.IsNationalStatistic = dlp.Description.NationalStatistic
	sdlp.DatasetLandingPage.IsTimeseries = dlp.Timeseries
	sdlp.ContactDetails = model.ContactDetails(dlp.Description.Contact)
	sdlp.DatasetLandingPage.ReleaseDate = dlp.Description.ReleaseDate
	sdlp.DatasetLandingPage.NextRelease = dlp.Description.NextRelease
	sdlp.DatasetLandingPage.Notes = dlp.Section.Markdown

	for _, bc := range bcs {
		sdlp.Page.Breadcrumb = append(sdlp.Page.Breadcrumb, model.TaxonomyNode{
			Title: bc.Description.Title,
			URI:   bc.URI,
		})
	}

	if len(sdlp.Page.Breadcrumb) > 0 {
		sdlp.DatasetLandingPage.ParentPath = sdlp.Page.Breadcrumb[len(sdlp.Page.Breadcrumb)-1].Title
	}

	for i, d := range ds {
		var dataset datasetLandingPageStatic.Dataset
		for _, value := range d.Downloads {
			dataset.URI = d.URI
			dataset.Downloads = append(dataset.Downloads, datasetLandingPageStatic.Download{
				URI:       value.File,
				Extension: strings.TrimPrefix(filepath.Ext(value.File), "."),
				Size:      value.Size,
			})
		}
		for _, value := range d.SupplementaryFiles {
			dataset.SupplementaryFiles = append(dataset.SupplementaryFiles, datasetLandingPageStatic.SupplementaryFile{
				Title:     value.Title,
				URI:       value.File,
				Extension: strings.TrimPrefix(filepath.Ext(value.File), "."),
				Size:      value.Size,
			})
		}
		dataset.Title = d.Description.Edition
		if len(d.Versions) > 0 {
			dataset.HasVersions = true
		}
		dataset.IsLast = i+1 == len(ds)
		sdlp.DatasetLandingPage.Datasets = append(sdlp.DatasetLandingPage.Datasets, dataset)
	}

	for _, value := range dlp.Alerts {
		switch value.Type {
		default:
			log.Debug("Unrecognised alert type", log.Data{"alert": value})
			fallthrough
		case "alert":
			sdlp.DatasetLandingPage.Notices = append(sdlp.DatasetLandingPage.Notices, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		case "correction":
			sdlp.DatasetLandingPage.Corrections = append(sdlp.DatasetLandingPage.Corrections, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		}
	}

	return sdlp
}
