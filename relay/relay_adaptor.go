package relay

import (
	"strconv"

	"github.com/xvyimu/TransitHub/constant"
	"github.com/xvyimu/TransitHub/relay/channel"
	"github.com/xvyimu/TransitHub/relay/channel/advancedcustom"
	"github.com/xvyimu/TransitHub/relay/channel/ali"
	"github.com/xvyimu/TransitHub/relay/channel/aws"
	"github.com/xvyimu/TransitHub/relay/channel/baidu"
	"github.com/xvyimu/TransitHub/relay/channel/baidu_v2"
	"github.com/xvyimu/TransitHub/relay/channel/claude"
	"github.com/xvyimu/TransitHub/relay/channel/cloudflare"
	"github.com/xvyimu/TransitHub/relay/channel/codex"
	"github.com/xvyimu/TransitHub/relay/channel/cohere"
	"github.com/xvyimu/TransitHub/relay/channel/coze"
	"github.com/xvyimu/TransitHub/relay/channel/deepseek"
	"github.com/xvyimu/TransitHub/relay/channel/dify"
	"github.com/xvyimu/TransitHub/relay/channel/gemini"
	"github.com/xvyimu/TransitHub/relay/channel/jimeng"
	"github.com/xvyimu/TransitHub/relay/channel/jina"
	"github.com/xvyimu/TransitHub/relay/channel/minimax"
	"github.com/xvyimu/TransitHub/relay/channel/mistral"
	"github.com/xvyimu/TransitHub/relay/channel/mokaai"
	"github.com/xvyimu/TransitHub/relay/channel/moonshot"
	"github.com/xvyimu/TransitHub/relay/channel/ollama"
	"github.com/xvyimu/TransitHub/relay/channel/openai"
	"github.com/xvyimu/TransitHub/relay/channel/palm"
	"github.com/xvyimu/TransitHub/relay/channel/perplexity"
	"github.com/xvyimu/TransitHub/relay/channel/replicate"
	"github.com/xvyimu/TransitHub/relay/channel/siliconflow"
	"github.com/xvyimu/TransitHub/relay/channel/submodel"
	taskali "github.com/xvyimu/TransitHub/relay/channel/task/ali"
	taskdoubao "github.com/xvyimu/TransitHub/relay/channel/task/doubao"
	taskGemini "github.com/xvyimu/TransitHub/relay/channel/task/gemini"
	"github.com/xvyimu/TransitHub/relay/channel/task/hailuo"
	taskjimeng "github.com/xvyimu/TransitHub/relay/channel/task/jimeng"
	"github.com/xvyimu/TransitHub/relay/channel/task/kling"
	tasksora "github.com/xvyimu/TransitHub/relay/channel/task/sora"
	"github.com/xvyimu/TransitHub/relay/channel/task/suno"
	taskvertex "github.com/xvyimu/TransitHub/relay/channel/task/vertex"
	taskVidu "github.com/xvyimu/TransitHub/relay/channel/task/vidu"
	"github.com/xvyimu/TransitHub/relay/channel/tencent"
	"github.com/xvyimu/TransitHub/relay/channel/vertex"
	"github.com/xvyimu/TransitHub/relay/channel/volcengine"
	"github.com/xvyimu/TransitHub/relay/channel/xai"
	"github.com/xvyimu/TransitHub/relay/channel/xunfei"
	"github.com/xvyimu/TransitHub/relay/channel/zhipu"
	"github.com/xvyimu/TransitHub/relay/channel/zhipu_4v"
	"github.com/gin-gonic/gin"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeZhipuV4:
		return &zhipu_4v.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	case constant.APITypeAws:
		return &aws.Adaptor{}
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	case constant.APITypeMokaAI:
		return &mokaai.Adaptor{}
	case constant.APITypeVolcEngine:
		return &volcengine.Adaptor{}
	case constant.APITypeBaiduV2:
		return &baidu_v2.Adaptor{}
	case constant.APITypeOpenRouter:
		return &openai.Adaptor{}
	case constant.APITypeXinference:
		return &openai.Adaptor{}
	case constant.APITypeXai:
		return &xai.Adaptor{}
	case constant.APITypeCoze:
		return &coze.Adaptor{}
	case constant.APITypeJimeng:
		return &jimeng.Adaptor{}
	case constant.APITypeMoonshot:
		return &moonshot.Adaptor{} // Moonshot uses Claude API
	case constant.APITypeSubmodel:
		return &submodel.Adaptor{}
	case constant.APITypeMiniMax:
		return &minimax.Adaptor{}
	case constant.APITypeReplicate:
		return &replicate.Adaptor{}
	case constant.APITypeCodex:
		return &codex.Adaptor{}
	case constant.APITypeAdvancedCustom:
		return &advancedcustom.Adaptor{}
	}
	return nil
}

func GetTaskPlatform(c *gin.Context) constant.TaskPlatform {
	channelType := c.GetInt("channel_type")
	if channelType > 0 {
		return constant.TaskPlatform(strconv.Itoa(channelType))
	}
	return constant.TaskPlatform(c.GetString("platform"))
}

func GetTaskAdaptor(platform constant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case constant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	if channelType, err := strconv.ParseInt(string(platform), 10, 64); err == nil {
		switch channelType {
		case constant.ChannelTypeAli:
			return &taskali.TaskAdaptor{}
		case constant.ChannelTypeKling:
			return &kling.TaskAdaptor{}
		case constant.ChannelTypeJimeng:
			return &taskjimeng.TaskAdaptor{}
		case constant.ChannelTypeVertexAi:
			return &taskvertex.TaskAdaptor{}
		case constant.ChannelTypeVidu:
			return &taskVidu.TaskAdaptor{}
		case constant.ChannelTypeDoubaoVideo, constant.ChannelTypeVolcEngine:
			return &taskdoubao.TaskAdaptor{}
		case constant.ChannelTypeSora, constant.ChannelTypeOpenAI:
			return &tasksora.TaskAdaptor{}
		case constant.ChannelTypeGemini:
			return &taskGemini.TaskAdaptor{}
		case constant.ChannelTypeMiniMax:
			return &hailuo.TaskAdaptor{}
		}
	}
	return nil
}
