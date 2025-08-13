package shared

type SearchResult struct {
	Title string
	URL   string
}

type SearchParams struct {
	Topic    string `json:"topic"`
	Duration string `json:"duration"`
	Sort_BY  string `json:"sort_by"`
}
type WithRefineOutput struct {
	RefineQuestions []string       `json:"refineQuestions" jsonschema_description:"A list of refining questions"`
	SearchResults   []SearchResult `json:"searchResults" jsonschema_description:"A list of search results that match query params"`
}
