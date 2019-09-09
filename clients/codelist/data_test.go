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

// EditionsListResults example entity used by unit tests.
var editionsListResults = EditionsListResults{
	TotalCount: 1,
	Offset:     0,
	Limit:      1,
	Count:      1,
	Items: []EditionsList{
		{
			Edition: "foo",
			Label:   "bar",
			Links: EditionsListLink{
				Self: &Link{
					Href: "/foo/bar",
					ID:   "1234567890",
				},
			},
		},
	},
}

// CodesResults example entity used by unit tests.
var codesResults = CodesResults{
	TotalCount: 1,
	Count:      1,
	Offset:     0,
	Limit:      1,
	Items: []Item{
		{
			ID:    "foo",
			Label: "bar",
			Links: CodeLinks{
				Self: Link{
					ID:   "1",
					Href: "/foo/bar",
				},
				Datasets: Link{
					ID:   "2",
					Href: "/datasets/foo/bar",
				},
				CodeLists: Link{
					ID:   "3",
					Href: "/codelists/foo/bar",
				},
			},
		},
	},
}
