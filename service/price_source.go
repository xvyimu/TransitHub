package service

import (
	"encoding/json"
	"strings"
)

// PriceSource identifies how a channel's model pricing should be fetched.
type PriceSource string

const (
	PriceSourceNewAPI     PriceSource = "newapi"
	PriceSourceSub2API    PriceSource = "sub2api"
	PriceSourceOpenRouter PriceSource = "openrouter"
	PriceSourceModelsDev  PriceSource = "models_dev"
	PriceSourceManual     PriceSource = "manual"
	PriceSourceUnknown    PriceSource = "unknown"
)

// GuessPriceSourceFromBaseURL is a cheap heuristic (no network).
func GuessPriceSourceFromBaseURL(baseURL string) PriceSource {
	u := strings.ToLower(strings.TrimSpace(baseURL))
	if u == "" {
		return PriceSourceUnknown
	}
	switch {
	case strings.Contains(u, "openrouter.ai"):
		return PriceSourceOpenRouter
	case strings.Contains(u, "models.dev"):
		return PriceSourceModelsDev
	case strings.Contains(u, "sub2api") || strings.Contains(u, "sub-2-api"):
		return PriceSourceSub2API
	case strings.Contains(u, "localhost") || strings.Contains(u, "127.0.0.1") ||
		strings.Contains(u, "new-api") || strings.Contains(u, "newapi"):
		return PriceSourceNewAPI
	default:
		// Most self-hosted OpenAI-compatible gateways in this stack are new-api forks.
		return PriceSourceUnknown
	}
}

// ClassifyPricingPayload inspects a successful JSON body and returns source + ratio-like data.
// Returns ok=false when payload is not a recognized pricing shape.
func ClassifyPricingPayload(body []byte) (source PriceSource, data map[string]any, ok bool) {
	// type1 envelope: {success, data: {model_ratio: {...}, ...}}
	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err == nil && len(env.Data) > 0 {
		var asMap map[string]any
		if err := json.Unmarshal(env.Data, &asMap); err == nil {
			if looksLikeRatioMap(asMap) {
				return PriceSourceNewAPI, asMap, true
			}
		}
		var asArr []any
		if err := json.Unmarshal(env.Data, &asArr); err == nil && len(asArr) > 0 {
			// type2 pricing list — caller already converts; mark as newapi
			if converted, err := convertPricingArrayToRatioData(env.Data); err == nil {
				return PriceSourceNewAPI, converted, true
			}
		}
	}

	// Bare ratio map
	var bare map[string]any
	if err := json.Unmarshal(body, &bare); err == nil && looksLikeRatioMap(bare) {
		return PriceSourceNewAPI, bare, true
	}

	// Sub2API / LiteLLM-style per-model cost tables
	if converted, ok := ConvertSub2APIStylePricing(body); ok {
		return PriceSourceSub2API, converted, true
	}

	return PriceSourceUnknown, nil, false
}

func looksLikeRatioMap(m map[string]any) bool {
	keys := []string{
		"model_ratio", "completion_ratio", "cache_ratio", "model_price",
		"create_cache_ratio", "image_ratio", "audio_ratio", "audio_completion_ratio",
	}
	for _, k := range keys {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

// ConvertSub2APIStylePricing maps common per-model USD cost tables into model_price
// (fixed price per 1M tokens input-style approximation) and optional model_ratio placeholders.
// This is best-effort; unknown shapes return ok=false.
func ConvertSub2APIStylePricing(body []byte) (map[string]any, bool) {
	// Shape A: { "claude-3-5": { "input_cost_per_token": 3e-6, "output_cost_per_token": 1.5e-5 }, ... }
	var nested map[string]map[string]any
	if err := json.Unmarshal(body, &nested); err == nil && len(nested) > 0 {
		price := map[string]any{}
		ratio := map[string]any{}
		comp := map[string]any{}
		count := 0
		for model, fields := range nested {
			if model == "" || fields == nil {
				continue
			}
			in, inOK := asFloat(fields["input_cost_per_token"])
			out, outOK := asFloat(fields["output_cost_per_token"])
			if !inOK && !outOK {
				// try alternate keys
				in, inOK = asFloat(fields["input_cost_per_million_tokens"])
				if inOK {
					in = in / 1_000_000
				}
				out, outOK = asFloat(fields["output_cost_per_million_tokens"])
				if outOK {
					out = out / 1_000_000
				}
			}
			if !inOK {
				continue
			}
			// Store USD per 1M input tokens as model_price when quota_type-like fixed pricing is intended.
			price[model] = in * 1_000_000
			if outOK && in > 0 {
				comp[model] = out / in
			}
			// Also expose a relative model_ratio scaled so 1e-6 input ≈ ratio 1 (common new-api baseline-ish).
			ratio[model] = (in * 1_000_000) / 2.0
			count++
		}
		if count == 0 {
			return nil, false
		}
		outMap := map[string]any{}
		if len(price) > 0 {
			outMap["model_price"] = price
		}
		if len(ratio) > 0 {
			outMap["model_ratio"] = ratio
		}
		if len(comp) > 0 {
			outMap["completion_ratio"] = comp
		}
		outMap["_adapter"] = "sub2api_style_cost_table"
		return outMap, true
	}

	// Shape B: { "data": { same nested } } without success flag
	var wrap struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err == nil && len(wrap.Data) > 0 {
		return ConvertSub2APIStylePricing(wrap.Data)
	}

	return nil, false
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// convertPricingArrayToRatioData is a minimal type2 converter for classification only.
func convertPricingArrayToRatioData(raw json.RawMessage) (map[string]any, error) {
	var items []struct {
		ModelName       string  `json:"model_name"`
		QuotaType       int     `json:"quota_type"`
		ModelRatio      float64 `json:"model_ratio"`
		ModelPrice      float64 `json:"model_price"`
		CompletionRatio float64 `json:"completion_ratio"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	modelRatio := map[string]any{}
	modelPrice := map[string]any{}
	completion := map[string]any{}
	for _, it := range items {
		if it.ModelName == "" {
			continue
		}
		if it.QuotaType == 1 {
			modelPrice[it.ModelName] = it.ModelPrice
		} else {
			modelRatio[it.ModelName] = it.ModelRatio
			completion[it.ModelName] = it.CompletionRatio
		}
	}
	out := map[string]any{}
	if len(modelRatio) > 0 {
		out["model_ratio"] = modelRatio
	}
	if len(modelPrice) > 0 {
		out["model_price"] = modelPrice
	}
	if len(completion) > 0 {
		out["completion_ratio"] = completion
	}
	return out, nil
}
