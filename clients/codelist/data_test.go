package codelist

// CodeListResults example entity used by unit tests.
var testCodeListResults = CodeListResults{
	Count:      1,
	Limit:      1,
	Offset:     0,
	TotalCount: 1,
	Items: []CodeList{
		{
			Links: CodeListLinks{
				Editions: &Link{
					ID:   "123",
					Href: "/123",
				},
				Self: &Link{
					ID:   "123",
					Href: "/123",
				},
			},
		},
	},
}

// DimensionValues example entity used by unit tests.
var testDimensionValues = DimensionValues{
	Items: []Item{
		{
			ID:    "123",
			Label: "Schwifty",
		},
	},
	NumberOfResults: 1,
}
