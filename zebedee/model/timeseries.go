package model

// TimeseriesPage is the root structure of the time series page.
type TimeseriesPage struct {
	Type        string            `json:"type"`
	URI         string            `json:"uri"`
	Description PageDescription   `json:"description"`
	Series      []TimeSeriesValue `json:"series"`
}

// TimeSeriesValue represents an individual time series entry.
type TimeSeriesValue struct {
	Name    string  `json:"name"`
	Y       float32 `json:"y"`
	StringY string  `json:"stringY"`
}
