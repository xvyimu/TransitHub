package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xvyimu/TransitHub/model"
	"github.com/xvyimu/TransitHub/service"

	"github.com/gin-gonic/gin"
)

type priceProbeRequest struct {
	ChannelID int    `json:"channel_id"`
	BaseURL   string `json:"base_url"`
	Endpoint  string `json:"endpoint"`
	Timeout   int    `json:"timeout"`
}

// ProbeUpstreamPricing classifies an upstream pricing endpoint without applying local ratios.
// Contract: never mutates Option / channel enablement (form: diff-only Apply policy).
func ProbeUpstreamPricing(c *gin.Context) {
	var req priceProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数格式错误"})
		return
	}
	if req.Timeout <= 0 {
		req.Timeout = 10
	}
	endpoint := req.Endpoint
	if endpoint == "" {
		endpoint = "/api/pricing"
	}
	if !strings.HasPrefix(endpoint, "/") && !strings.HasPrefix(endpoint, "http") {
		endpoint = "/" + endpoint
	}

	baseURL := strings.TrimRight(strings.TrimSpace(req.BaseURL), "/")
	channelName := ""
	if req.ChannelID > 0 {
		ch, err := model.GetChannelById(req.ChannelID, false)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "渠道不存在"})
			return
		}
		channelName = ch.Name
		if baseURL == "" {
			baseURL = strings.TrimRight(ch.GetBaseURL(), "/")
		}
	}
	if baseURL == "" || !strings.HasPrefix(baseURL, "http") {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无有效 base_url"})
		return
	}

	guess := service.GuessPriceSourceFromBaseURL(baseURL)
	fullURL := baseURL + endpoint
	if strings.HasPrefix(endpoint, "http") {
		fullURL = endpoint
	}

	client := &http.Client{Timeout: time.Duration(req.Timeout) * time.Second}
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, fullURL, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"channel_id":    req.ChannelID,
			"channel_name":  channelName,
			"base_url":      baseURL,
			"endpoint":      endpoint,
			"price_source":  guess,
			"ok":            false,
			"message":       err.Error(),
			"probed_at":     time.Now().Unix(),
			"apply_mutated": false,
		}})
		return
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"channel_id":    req.ChannelID,
			"channel_name":  channelName,
			"base_url":      baseURL,
			"endpoint":      endpoint,
			"price_source":  guess,
			"ok":            false,
			"message":       err.Error(),
			"probed_at":     time.Now().Unix(),
			"apply_mutated": false,
		}})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"channel_id":    req.ChannelID,
			"channel_name":  channelName,
			"base_url":      baseURL,
			"endpoint":      endpoint,
			"price_source":  guess,
			"ok":            false,
			"http_status":   resp.StatusCode,
			"message":       resp.Status,
			"probed_at":     time.Now().Unix(),
			"apply_mutated": false,
		}})
		return
	}

	src, data, ok := service.ClassifyPricingPayload(body)
	if !ok {
		// try sub2api style even if envelope failed
		if converted, ok2 := service.ConvertSub2APIStylePricing(body); ok2 {
			src, data, ok = service.PriceSourceSub2API, converted, true
		}
	}
	if !ok {
		msg := "无法识别为 new-api 定价或 sub2api 成本表"
		if guess == service.PriceSourceSub2API {
			msg = "疑似 sub2api 主机但响应无法映射，已标记 unknown（不禁用渠道）"
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"channel_id":    req.ChannelID,
			"channel_name":  channelName,
			"base_url":      baseURL,
			"endpoint":      endpoint,
			"price_source":  service.PriceSourceUnknown,
			"guess":         guess,
			"ok":            false,
			"message":       msg,
			"probed_at":     time.Now().Unix(),
			"apply_mutated": false,
		}})
		return
	}

	modelCount := 0
	for _, key := range []string{"model_ratio", "model_price"} {
		if m, ok := data[key].(map[string]any); ok {
			modelCount += len(m)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"channel_id":    req.ChannelID,
		"channel_name":  channelName,
		"base_url":      baseURL,
		"endpoint":      endpoint,
		"price_source":  src,
		"guess":         guess,
		"ok":            true,
		"model_count":   modelCount,
		"has_adapter":   data["_adapter"] != nil,
		"message":       "",
		"probed_at":     time.Now().Unix(),
		"apply_mutated": false,
		// Never auto-apply: expose field keys only for operators.
		"ratio_fields": ratioFieldNames(data),
	}})
}

func ratioFieldNames(data map[string]any) []string {
	out := make([]string, 0, len(data))
	for k := range data {
		if strings.HasPrefix(k, "_") {
			continue
		}
		out = append(out, k)
	}
	return out
}

// RunRatioSyncSnapshot fetches upstream ratios and writes a JSON snapshot without applying Option.
func RunRatioSyncSnapshot(c *gin.Context) {
	// Reuse fetch body shape.
	var req struct {
		ChannelIDs []int64 `json:"channel_ids"`
		Timeout    int     `json:"timeout"`
		Note       string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数格式错误"})
		return
	}
	if req.Timeout <= 0 {
		req.Timeout = 10
	}

	// Call internal fetch by synthesizing the same handler request is heavy;
	// instead snapshot metadata + lightweight probe per channel.
	type item struct {
		ChannelID   int64               `json:"channel_id"`
		Name        string              `json:"name"`
		BaseURL     string              `json:"base_url"`
		PriceSource service.PriceSource `json:"price_source"`
		OK          bool                `json:"ok"`
		Message     string              `json:"message,omitempty"`
		ModelCount  int                 `json:"model_count,omitempty"`
	}
	items := make([]item, 0, len(req.ChannelIDs))
	client := &http.Client{Timeout: time.Duration(req.Timeout) * time.Second}

	for _, id := range req.ChannelIDs {
		ch, err := model.GetChannelById(int(id), false)
		if err != nil {
			items = append(items, item{ChannelID: id, OK: false, Message: "channel not found"})
			continue
		}
		base := strings.TrimRight(ch.GetBaseURL(), "/")
		it := item{
			ChannelID:   id,
			Name:        ch.Name,
			BaseURL:     base,
			PriceSource: service.GuessPriceSourceFromBaseURL(base),
		}
		if base == "" || !strings.HasPrefix(base, "http") {
			it.OK = false
			it.Message = "empty base_url"
			items = append(items, it)
			continue
		}
		url := base + "/api/pricing"
		httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		if err != nil {
			it.OK = false
			it.Message = err.Error()
			items = append(items, it)
			continue
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			// fallback ratio_config then sub2 style raw
			it.OK = false
			it.Message = err.Error()
			items = append(items, it)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			// try ratio_config
			url2 := base + "/api/ratio_config"
			httpReq2, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url2, nil)
			if httpReq2 != nil {
				if resp2, err2 := client.Do(httpReq2); err2 == nil {
					body, _ = io.ReadAll(io.LimitReader(resp2.Body, 10<<20))
					resp2.Body.Close()
					if resp2.StatusCode == http.StatusOK {
						if src, data, ok := service.ClassifyPricingPayload(body); ok {
							it.OK = true
							it.PriceSource = src
							for _, key := range []string{"model_ratio", "model_price"} {
								if m, ok := data[key].(map[string]any); ok {
									it.ModelCount += len(m)
								}
							}
							items = append(items, it)
							continue
						}
					}
				}
			}
			it.OK = false
			it.Message = resp.Status
			items = append(items, it)
			continue
		}
		src, data, ok := service.ClassifyPricingPayload(body)
		if !ok {
			if converted, ok2 := service.ConvertSub2APIStylePricing(body); ok2 {
				src, data, ok = service.PriceSourceSub2API, converted, true
			}
		}
		if !ok {
			it.OK = false
			it.Message = "unrecognized pricing payload"
			items = append(items, it)
			continue
		}
		it.OK = true
		it.PriceSource = src
		for _, key := range []string{"model_ratio", "model_price"} {
			if m, ok := data[key].(map[string]any); ok {
				it.ModelCount += len(m)
			}
		}
		items = append(items, it)
	}

	dir := `D:\newapi\backups\ratio-snapshots`
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "创建快照目录失败: " + err.Error()})
		return
	}
	stamp := time.Now().Format("20060102-150405")
	path := filepath.Join(dir, fmt.Sprintf("ratio-snapshot-%s.json", stamp))
	payload := gin.H{
		"created_at":    time.Now().Format(time.RFC3339),
		"note":          req.Note,
		"apply_mutated": false,
		"policy":        "snapshot_only_no_auto_apply",
		"live_hint":     "use curl /livez for binary version; this file is pricing probe snapshot only",
		"items":         items,
	}
	raw, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "写入快照失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"path":          path,
			"item_count":    len(items),
			"apply_mutated": false,
			"items":         items,
		},
	})
}

// ListRatioSyncSnapshots returns recent snapshot files (metadata only).
func ListRatioSyncSnapshots(c *gin.Context) {
	dirs := []string{
		`D:\newapi\backups\ratio-snapshots`,
	}
	type meta struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Size    int64  `json:"size"`
		ModTime int64  `json:"mod_time"`
	}
	out := make([]meta, 0)
	seen := map[string]bool{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasPrefix(e.Name(), "ratio-snapshot-") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			p := filepath.Join(dir, e.Name())
			if seen[e.Name()] {
				continue
			}
			seen[e.Name()] = true
			out = append(out, meta{
				Name:    e.Name(),
				Path:    p,
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": out, "apply_mutated": false})
}
