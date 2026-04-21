package service

import (
	"encoding/json"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
)

func extractResponsesOutputText(resp *apicompat.ResponsesResponse) string {
	if resp == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	for _, item := range resp.Output {
		for _, content := range item.Content {
			switch content.Type {
			case "output_text", "text", "input_text":
				if text := strings.TrimSpace(content.Text); text != "" {
					parts = append(parts, text)
				}
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func extractAnthropicResponseText(resp *apicompat.AnthropicResponse) string {
	if resp == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	for _, block := range resp.Content {
		if text := strings.TrimSpace(block.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func extractChatCompletionsAssistantText(resp *apicompat.ChatCompletionsResponse) string {
	if resp == nil {
		return ""
	}
	parts := make([]string, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		if text := extractChatMessageText(choice.Message); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func extractChatMessageText(msg apicompat.ChatMessage) string {
	if len(msg.Content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(msg.Content, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var parts []apicompat.ChatContentPart
	if err := json.Unmarshal(msg.Content, &parts); err == nil {
		texts := make([]string, 0, len(parts))
		for _, part := range parts {
			if part.Type == "text" {
				if text := strings.TrimSpace(part.Text); text != "" {
					texts = append(texts, text)
				}
			}
		}
		return strings.TrimSpace(strings.Join(texts, "\n"))
	}
	return strings.TrimSpace(string(msg.Content))
}

func gjsonResponseFromSSEPayload(payload []byte) *apicompat.ResponsesResponse {
	if len(payload) == 0 {
		return nil
	}
	var wrapper struct {
		Response *apicompat.ResponsesResponse `json:"response"`
		Item     *apicompat.ResponsesOutput   `json:"item"`
	}
	if err := json.Unmarshal(payload, &wrapper); err != nil {
		return nil
	}
	if wrapper.Response != nil {
		return wrapper.Response
	}
	if wrapper.Item != nil {
		return &apicompat.ResponsesResponse{Output: []apicompat.ResponsesOutput{*wrapper.Item}}
	}
	return nil
}
