package service

import "strings"

// resolveOpenAIForwardModel 解析 OpenAI 兼容转发使用的模型。
// messagesDispatchMappedModel 是调用方已为 /v1/messages 解析的显式调度结果；
// 普通 OpenAI 请求必须传空，避免将分组配置作为通用模型兜底。
func resolveOpenAIForwardModel(account *Account, requestedModel, messagesDispatchMappedModel string) string {
	messagesDispatchMappedModel = strings.TrimSpace(messagesDispatchMappedModel)
	if account == nil {
		if messagesDispatchMappedModel != "" {
			return messagesDispatchMappedModel
		}
		return requestedModel
	}

	if fallbackModel, ok := resolveOpenAIGPT55FallbackModel(account, requestedModel); ok {
		requestedModel = fallbackModel
	}

	mappedModel, matched := account.ResolveMappedModel(requestedModel)
	if !matched && messagesDispatchMappedModel != "" {
		return messagesDispatchMappedModel
	}
	return mappedModel
}

func resolveOpenAIGPT55FallbackModel(account *Account, requestedModel string) (string, bool) {
	if account == nil || account.Platform != PlatformOpenAI || account.Type != AccountTypeOAuth {
		return "", false
	}
	if !openAIGPT55FallbackEnabled(account) {
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

func openAIGPT55FallbackEnabled(account *Account) bool {
	if account == nil || account.Credentials == nil {
		return false
	}
	switch v := account.Credentials["gpt55_fallback_enabled"].(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(strings.TrimSpace(v), "true") || strings.TrimSpace(v) == "1"
	default:
		return false
	}
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
	switch lowered {
	case "gpt-5.5":
		return normalizedOpenAIGPT55Model{
			original: prefix + "gpt-5.5",
			fallback: prefix + "gpt-5.4",
		}, true
	case "gpt-5.5-openai-compact":
		return normalizedOpenAIGPT55Model{
			original: prefix + "gpt-5.5-openai-compact",
			fallback: prefix + "gpt-5.4-openai-compact",
		}, true
	default:
		return normalizedOpenAIGPT55Model{}, false
	}
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

// openAIOAuthForeignModelPrefixes 列出明确属于其他厂商家族的模型名前缀。
// Codex 上游不可能服务这些模型：转发阶段 normalizeOpenAIModelForUpstream
// 对未知模型原样透传，上游必然返回不可重试的 400。
//
// 采用保守黑名单而非 Codex 模型白名单：未知/自定义别名保持「允许」，
// 以兼容渠道级模型映射等「账号选定之后才改写模型名」的部署方式
// （调度过滤看到的是改写前的原始模型名）。前缀分类的先例见
// ResolveThinkingProtocol（thinking_protocol.go）。
var openAIOAuthForeignModelPrefixes = []string{
	"deepseek-",
	"glm-",
	"kimi-",
	"moonshot-",
	"qwen-",
	"qwen2-",
	"qwen3-",
	"qwen4-",
	"qwq-",
	"minimax-",
	"gemini-",
	"gemma-",
	"grok-",
	"doubao-",
	"hunyuan-",
	"llama-",
	"llama2-",
	"llama3-",
	"meta-llama",
	"mistral-",
	"mixtral-",
	"baichuan-",
	"ernie-",
	"step-",
	"seed-",
	"yi-",
}

// isOpenAIOAuthServableModel 判断「空 model_mapping 的 OpenAI OAuth 账号」能否
// 服务请求模型。空映射默认仍是「允许」，仅排除明确属于其他厂商家族的模型
// （deepseek-*/glm-* 等）——这类请求原样透传必然被 Codex 上游以不可重试的
// 400 拒绝，且不触发 failover，应在调度阶段就跳过该账号，把请求让给
// 显式声明支持该模型的账号（#3662）。
func isOpenAIOAuthServableModel(requestedModel string) bool {
	model := strings.ToLower(lastOpenAIModelSegment(requestedModel))
	if model == "" {
		return true // 空模型交由上层必填校验处理
	}
	for _, prefix := range openAIOAuthForeignModelPrefixes {
		if strings.HasPrefix(model, prefix) {
			return false
		}
	}
	return true
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
