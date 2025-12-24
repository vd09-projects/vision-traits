package traits

// JSON Schema (draft-ish) object to pass into Ollama `format`.
// Keep it tight: no extra keys.
func TraitsJSONSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]interface{}{
			"global_confidence": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 100},
			"traits": map[string]interface{}{
				"type": "object",
				"additionalProperties": map[string]interface{}{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]interface{}{
						"summary":    map[string]interface{}{"type": "string"},
						"signals":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
						"confidence": map[string]interface{}{"type": "integer", "minimum": 0, "maximum": 100},
					},
					"required": []string{"summary", "signals", "confidence"},
				},
			},
			"notes": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
		},
		"required": []string{"global_confidence", "traits", "notes"},
	}
}
