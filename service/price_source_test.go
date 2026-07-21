package service

import (
	"encoding/json"
	"testing"
)

func TestGuessPriceSourceFromBaseURL(t *testing.T) {
	cases := map[string]PriceSource{
		"https://openrouter.ai/api":     PriceSourceOpenRouter,
		"https://models.dev":            PriceSourceModelsDev,
		"http://127.0.0.1:3000":         PriceSourceNewAPI,
		"https://my-sub2api.example":    PriceSourceSub2API,
		"":                              PriceSourceUnknown,
		"https://relay.example.com":     PriceSourceUnknown,
	}
	for in, want := range cases {
		if got := GuessPriceSourceFromBaseURL(in); got != want {
			t.Fatalf("Guess(%q)=%q want %q", in, got, want)
		}
	}
}

func TestClassifyNewAPIType1(t *testing.T) {
	body := []byte(`{"success":true,"data":{"model_ratio":{"gpt-4o-mini":0.07},"completion_ratio":{"gpt-4o-mini":4}}}`)
	src, data, ok := ClassifyPricingPayload(body)
	if !ok || src != PriceSourceNewAPI {
		t.Fatalf("ok=%v src=%s", ok, src)
	}
	if _, has := data["model_ratio"]; !has {
		t.Fatal("missing model_ratio")
	}
}

func TestConvertSub2APIStylePricing(t *testing.T) {
	body := []byte(`{
		"claude-3-5-sonnet": {
			"input_cost_per_token": 0.000003,
			"output_cost_per_token": 0.000015
		}
	}`)
	data, ok := ConvertSub2APIStylePricing(body)
	if !ok {
		t.Fatal("expected ok")
	}
	price := data["model_price"].(map[string]any)
	if price["claude-3-5-sonnet"] == nil {
		t.Fatal("missing model price")
	}
	comp := data["completion_ratio"].(map[string]any)
	if comp["claude-3-5-sonnet"] == nil {
		t.Fatal("missing completion ratio")
	}
}

func TestClassifySub2API(t *testing.T) {
	body := []byte(`{"deepseek-v3":{"input_cost_per_token":1e-7,"output_cost_per_token":2e-7}}`)
	src, data, ok := ClassifyPricingPayload(body)
	if !ok || src != PriceSourceSub2API {
		t.Fatalf("ok=%v src=%s data=%v", ok, src, data)
	}
}

func TestClassifyRejectsGarbage(t *testing.T) {
	_, _, ok := ClassifyPricingPayload([]byte(`{"hello":"world"}`))
	if ok {
		t.Fatal("expected reject")
	}
	_ = json.RawMessage{}
}
