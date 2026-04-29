package service

import "strings"

// resolveOpenAIForwardModel determines the upstream model for OpenAI-compatible
// forwarding. Group-level default mapping only applies when the account itself
// did not match any explicit model_mapping rule.
func resolveOpenAIForwardModel(account *Account, requestedModel, defaultMappedModel string) string {
	if account == nil {
		if defaultMappedModel != "" {
			return defaultMappedModel
		}
		return requestedModel
	}

	if fallbackModel, ok := resolveOpenAIGPT55FallbackModel(account, requestedModel); ok {
		requestedModel = fallbackModel
	}

	mappedModel, matched := account.ResolveMappedModel(requestedModel)
	if !matched && defaultMappedModel != "" && !isExplicitCodexModel(requestedModel) {
		return defaultMappedModel
	}
	return mappedModel
}

func resolveOpenAIGPT55FallbackModel(account *Account, requestedModel string) (string, bool) {
	if account == nil || account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth {
		return "", false
	}

	normalized, ok := normalizeOpenAIGPT55RequestedModel(requestedModel)
	if !ok {
		return "", false
	}
	if openAIAccountAdvertisesModel(account, normalized.original) {
		return "", false
	}
	if openAIAccountAdvertisesModel(account, normalized.fallback) {
		return normalized.fallback, true
	}
	return "", false
}

type normalizedOpenAIGPT55Model struct {
	original string
	fallback string
}

func normalizeOpenAIGPT55RequestedModel(requestedModel string) (normalizedOpenAIGPT55Model, bool) {
	trimmed := strings.TrimSpace(requestedModel)
	if trimmed == "" {
		return normalizedOpenAIGPT55Model{}, false
	}

	prefix := ""
	modelPart := trimmed
	if slash := strings.LastIndex(trimmed, "/"); slash >= 0 {
		prefix = trimmed[:slash+1]
		modelPart = trimmed[slash+1:]
	}

	lowered := strings.ToLower(modelPart)
	if lowered == "gpt-5.5" {
		return normalizedOpenAIGPT55Model{
			original: prefix + "gpt-5.5",
			fallback: prefix + "gpt-5.4",
		}, true
	}
	if lowered == "gpt-5.5-openai-compact" {
		return normalizedOpenAIGPT55Model{
			original: prefix + "gpt-5.5-openai-compact",
			fallback: prefix + "gpt-5.4-openai-compact",
		}, true
	}
	return normalizedOpenAIGPT55Model{}, false
}

func openAIAccountAdvertisesModel(account *Account, model string) bool {
	if account == nil {
		return false
	}
	snapshot := account.GetSupportedModelsSnapshot()
	if len(snapshot) > 0 {
		lookup := strings.TrimSpace(strings.TrimPrefix(model, "openai/"))
		for _, item := range snapshot {
			if item == lookup {
				return true
			}
		}
		return false
	}

	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		return false
	}
	if mappingSupportsRequestedModel(mapping, model) {
		return true
	}
	normalized := strings.TrimSpace(strings.TrimPrefix(model, "openai/"))
	return normalized != model && mappingSupportsRequestedModel(mapping, normalized)
}

func isExplicitCodexModel(model string) bool {
	model = strings.TrimSpace(model)
	if model == "" {
		return false
	}
	if strings.Contains(model, "/") {
		parts := strings.Split(model, "/")
		model = parts[len(parts)-1]
	}
	model = strings.ToLower(strings.TrimSpace(model))
	if getNormalizedCodexModel(model) != "" {
		return true
	}
	if strings.HasSuffix(model, "-openai-compact") {
		base := strings.TrimSuffix(model, "-openai-compact")
		return getNormalizedCodexModel(base) != ""
	}
	return false
}

// resolveOpenAICompactForwardModel determines the compact-only upstream model
// for /responses/compact requests. It never affects normal /responses traffic.
// When no compact-specific mapping matches, the input model is returned as-is.
func resolveOpenAICompactForwardModel(account *Account, model string) string {
	trimmedModel := strings.TrimSpace(model)
	if trimmedModel == "" || account == nil {
		return trimmedModel
	}

	mappedModel, matched := account.ResolveCompactMappedModel(trimmedModel)
	if !matched {
		return trimmedModel
	}
	if trimmedMapped := strings.TrimSpace(mappedModel); trimmedMapped != "" {
		return trimmedMapped
	}
	return trimmedModel
}
