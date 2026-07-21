package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/constant"
	"github.com/xvyimu/TransitHub/dto"
	"github.com/xvyimu/TransitHub/model"
	relaychannel "github.com/xvyimu/TransitHub/relay/channel"
	relaycommon "github.com/xvyimu/TransitHub/relay/common"

	"gorm.io/gorm"
)

type OpenAIModel struct {
	ID         string         `json:"id"`
	Object     string         `json:"object"`
	Created    int64          `json:"created"`
	OwnedBy    string         `json:"owned_by"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Permission []struct {
		ID                 string `json:"id"`
		Object             string `json:"object"`
		Created            int64  `json:"created"`
		AllowCreateEngine  bool   `json:"allow_create_engine"`
		AllowSampling      bool   `json:"allow_sampling"`
		AllowLogprobs      bool   `json:"allow_logprobs"`
		AllowSearchIndices bool   `json:"allow_search_indices"`
		AllowView          bool   `json:"allow_view"`
		AllowFineTuning    bool   `json:"allow_fine_tuning"`
		Organization       string `json:"organization"`
		Group              string `json:"group"`
		IsBlocking         bool   `json:"is_blocking"`
	} `json:"permission"`
	Root   string `json:"root"`
	Parent string `json:"parent"`
}

type OpenAIModelsResponse struct {
	Data    []OpenAIModel `json:"data"`
	Success bool          `json:"success"`
}

func parseStatusFilter(statusParam string) int {
	switch strings.ToLower(statusParam) {
	case "enabled", "1":
		return common.ChannelStatusEnabled
	case "disabled", "0":
		return 0
	default:
		return -1
	}
}

func clearChannelInfo(channel *model.Channel) {
	if channel.ChannelInfo.IsMultiKey {
		channel.ChannelInfo.MultiKeyDisabledReason = nil
		channel.ChannelInfo.MultiKeyDisabledTime = nil
	}
}

func applyChannelStatusFilter(query *gorm.DB, statusFilter int) *gorm.DB {
	if statusFilter == common.ChannelStatusEnabled {
		return query.Where("status = ?", common.ChannelStatusEnabled)
	}
	if statusFilter == 0 {
		return query.Where("status != ?", common.ChannelStatusEnabled)
	}
	return query
}

func buildChannelListQuery(group string, statusFilter int, typeFilter int) *gorm.DB {
	query := model.DB.Model(&model.Channel{})
	query = model.ApplyChannelGroupFilter(query, group)
	query = applyChannelStatusFilter(query, statusFilter)
	if typeFilter >= 0 {
		query = query.Where("type = ?", typeFilter)
	}
	return query
}

func buildFetchModelsHeaders(channel *model.Channel, key string) (http.Header, error) {
	var headers http.Header
	switch channel.Type {
	case constant.ChannelTypeAnthropic:
		headers = GetClaudeAuthHeader(key)
	default:
		headers = GetAuthHeader(key)
	}

	if err := applyFetchModelsHeaderOverrides(channel, key, headers); err != nil {
		return nil, err
	}
	return headers, nil
}

func applyFetchModelsHeaderOverrides(channel *model.Channel, key string, headers http.Header) error {
	info := &relaycommon.RelayInfo{
		IsChannelTest: true,
		ChannelMeta: &relaycommon.ChannelMeta{
			ApiKey:          key,
			HeadersOverride: channel.GetHeaderOverride(),
		},
	}
	overrides, err := relaychannel.ResolveHeaderOverride(info, nil)
	if err != nil {
		return err
	}
	for name, value := range overrides {
		headers.Set(name, value)
	}

	return nil
}

func validateTwoFactorAuth(twoFA *model.TwoFA, code string) bool {
	// 尝试验证TOTP
	if cleanCode, err := common.ValidateNumericCode(code); err == nil {
		if isValid, _ := twoFA.ValidateTOTPAndUpdateUsage(cleanCode); isValid {
			return true
		}
	}

	// 尝试验证备用码
	if isValid, err := twoFA.ValidateBackupCodeAndUpdateUsage(code); err == nil && isValid {
		return true
	}

	return false
}

// validateChannel 通用的渠道校验函数
func validateChannel(channel *model.Channel, isAdd bool) error {
	if channel == nil {
		return fmt.Errorf("channel cannot be empty")
	}

	// 校验 channel settings
	if err := channel.ValidateSettings(); err != nil {
		return fmt.Errorf("渠道额外设置[channel setting] 格式错误：%s", err.Error())
	}

	// 如果是添加操作，检查 channel 和 key 是否为空
	if isAdd {
		if channel.Key == "" {
			return fmt.Errorf("channel cannot be empty")
		}

		// 检查模型名称长度是否超过 255
		for _, m := range channel.GetModels() {
			if len(m) > 255 {
				return fmt.Errorf("模型名称过长: %s", m)
			}
		}
	}

	// VertexAI 特殊校验
	if channel.Type == constant.ChannelTypeVertexAi {
		if channel.Other == "" {
			return fmt.Errorf("部署地区不能为空")
		}

		regionMap, err := common.StrToMap(channel.Other)
		if err != nil {
			return fmt.Errorf("部署地区必须是标准的Json格式，例如{\"default\": \"us-central1\", \"region2\": \"us-east1\"}")
		}

		if regionMap["default"] == nil {
			return fmt.Errorf("部署地区必须包含default字段")
		}
	}

	// Codex OAuth key validation (optional, only when JSON object is provided)
	if channel.Type == constant.ChannelTypeCodex {
		trimmedKey := strings.TrimSpace(channel.Key)
		if isAdd || trimmedKey != "" {
			if !strings.HasPrefix(trimmedKey, "{") {
				return fmt.Errorf("Codex key must be a valid JSON object")
			}
			var keyMap map[string]any
			if err := common.Unmarshal([]byte(trimmedKey), &keyMap); err != nil {
				return fmt.Errorf("Codex key must be a valid JSON object")
			}
			if v, ok := keyMap["access_token"]; !ok || v == nil || strings.TrimSpace(fmt.Sprintf("%v", v)) == "" {
				return fmt.Errorf("Codex key JSON must include access_token")
			}
			if v, ok := keyMap["account_id"]; !ok || v == nil || strings.TrimSpace(fmt.Sprintf("%v", v)) == "" {
				return fmt.Errorf("Codex key JSON must include account_id")
			}
		}
	}

	return nil
}

func getVertexArrayKeys(keys string) ([]string, error) {
	if keys == "" {
		return nil, nil
	}
	var keyArray []interface{}
	err := common.Unmarshal([]byte(keys), &keyArray)
	if err != nil {
		return nil, fmt.Errorf("批量添加 Vertex AI 必须使用标准的JsonArray格式，例如[{key1}, {key2}...]，请检查输入: %w", err)
	}
	cleanKeys := make([]string, 0, len(keyArray))
	for _, key := range keyArray {
		var keyStr string
		switch v := key.(type) {
		case string:
			keyStr = strings.TrimSpace(v)
		default:
			bytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("Vertex AI key JSON 编码失败: %w", err)
			}
			keyStr = string(bytes)
		}
		if keyStr != "" {
			cleanKeys = append(cleanKeys, keyStr)
		}
	}
	if len(cleanKeys) == 0 {
		return nil, fmt.Errorf("批量添加 Vertex AI 的 keys 不能为空")
	}
	return cleanKeys, nil
}

func isManageableChannelStatus(status int) bool {
	return status == common.ChannelStatusEnabled || status == common.ChannelStatusManuallyDisabled
}

// equalStringPtr 比较两个 *string 是否相等（均为 nil 视为相等）。
func equalStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

type fetchModelsRequest struct {
	ChannelID      int     `json:"channel_id"`
	BaseURL        *string `json:"base_url"`
	Type           int     `json:"type"`
	Key            string  `json:"key"`
	AdvancedCustom *string `json:"advanced_custom"`
	HeaderOverride *string `json:"header_override"`
	Proxy          *string `json:"proxy"`
}

func buildAdvancedCustomModelPreviewChannel(req fetchModelsRequest) (*model.Channel, error) {
	var channel *model.Channel
	if req.ChannelID > 0 {
		savedChannel, err := model.GetChannelById(req.ChannelID, true)
		if err != nil {
			return nil, err
		}
		if savedChannel.Type != constant.ChannelTypeAdvancedCustom {
			return nil, fmt.Errorf("channel %d is not an advanced custom channel", req.ChannelID)
		}
		channel = savedChannel
	} else {
		key := strings.TrimSpace(req.Key)
		if key != "" {
			key = strings.Split(key, "\n")[0]
		}
		channel = &model.Channel{
			Type: req.Type,
			Key:  key,
		}
	}

	if channel.Type != constant.ChannelTypeAdvancedCustom {
		return nil, fmt.Errorf("channel type must be advanced custom")
	}
	if req.BaseURL != nil {
		baseURL := strings.TrimSpace(*req.BaseURL)
		channel.BaseURL = &baseURL
	}

	settings := channel.GetOtherSettings()
	if req.AdvancedCustom != nil {
		rawConfig := strings.TrimSpace(*req.AdvancedCustom)
		if rawConfig == "" {
			return nil, fmt.Errorf("advanced_custom is required")
		}
		var config dto.AdvancedCustomConfig
		if err := common.UnmarshalJsonStr(rawConfig, &config); err != nil {
			return nil, err
		}
		settings.AdvancedCustom = &config
	} else if req.ChannelID <= 0 {
		return nil, fmt.Errorf("advanced_custom is required")
	}
	channel.SetOtherSettings(settings)

	if req.HeaderOverride != nil {
		rawHeaderOverride := strings.TrimSpace(*req.HeaderOverride)
		if rawHeaderOverride != "" {
			var headerOverride map[string]any
			if err := common.UnmarshalJsonStr(rawHeaderOverride, &headerOverride); err != nil {
				return nil, fmt.Errorf("header_override must be a JSON object: %w", err)
			}
		}
		channel.HeaderOverride = &rawHeaderOverride
	}
	if req.Proxy != nil {
		channelSettings := channel.GetSetting()
		channelSettings.Proxy = strings.TrimSpace(*req.Proxy)
		channel.SetSetting(channelSettings)
	}

	if err := validateChannel(channel, false); err != nil {
		return nil, err
	}
	return channel, nil
}

func multiKeyActionRequiresSensitiveWrite(action string) bool {
	return action == "delete_key" || action == "delete_disabled_keys"
}

// OllamaPullModel 拉取 Ollama 模型
