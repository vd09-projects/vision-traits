package traits

type ExtractedTraits struct {
	GlobalConfidence int                            `json:"global_confidence"`
	Traits           map[string]TraitCategoryResult `json:"traits"`
	Notes            []string                       `json:"notes"`
}

type TraitCategoryResult struct {
	Summary       string              `json:"summary"`
	Signals       []string            `json:"signals"`
	SignalsByKey  map[string][]string `json:"signals_by_key,omitempty"`
	Confidence    int                 `json:"confidence"`
}
