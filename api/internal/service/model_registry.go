package service

// ModelTier represents the capability tier for model selection.
type ModelTier int

const (
	TierStandard  ModelTier = iota // Full-capability model for chat, entity extraction
	TierFast                       // Lightweight model for intent classification, autocomplete
	TierEmbedding                  // Embedding model
)

// ModelRegistry maps task tiers to specific model names.
type ModelRegistry struct {
	models map[ModelTier]string
}

// NewModelRegistry creates a registry. If fast is empty, it falls back to standard.
func NewModelRegistry(standard, fast, embedding string) *ModelRegistry {
	if fast == "" {
		fast = standard
	}
	return &ModelRegistry{
		models: map[ModelTier]string{
			TierStandard:  standard,
			TierFast:      fast,
			TierEmbedding: embedding,
		},
	}
}

// Model returns the model name for the given tier, falling back to standard.
func (r *ModelRegistry) Model(tier ModelTier) string {
	if m, ok := r.models[tier]; ok && m != "" {
		return m
	}
	return r.models[TierStandard]
}
