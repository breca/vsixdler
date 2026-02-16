package marketplace

// API request types

type QueryRequest struct {
	Filters []Filter `json:"filters"`
	Flags   int      `json:"flags"`
}

type Filter struct {
	Criteria []Criterion `json:"criteria"`
}

type Criterion struct {
	FilterType int    `json:"filterType"`
	Value      string `json:"value"`
}

// API response types

type QueryResponse struct {
	Results []QueryResult `json:"results"`
}

type QueryResult struct {
	Extensions []ExtensionResult `json:"extensions"`
}

type ExtensionResult struct {
	Publisher Publisher        `json:"publisher"`
	Name      string          `json:"extensionName"`
	Versions  []VersionResult `json:"versions"`
}

type Publisher struct {
	Name string `json:"publisherName"`
}

type VersionResult struct {
	Version          string           `json:"version"`
	TargetPlatform   string           `json:"targetPlatform,omitempty"`
	Files            []FileResult     `json:"files"`
	Properties       []PropertyResult `json:"properties"`
}

type FileResult struct {
	AssetType string `json:"assetType"`
	Source    string `json:"source"`
}

type PropertyResult struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Flags for the gallery API
const (
	// FlagsLatest returns only the latest version per target platform.
	FlagsLatest = 0x3D6 // 982

	// FlagsAllVersions returns all versions so we can find a pinned one.
	FlagsAllVersions = 0x1D6 // 470
)
