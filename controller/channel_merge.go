package controller

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type channelMergeBody struct {
	Ids       []int `json:"ids"`
	PrimaryId int   `json:"primary_id"`
}

// ListDuplicateChannels returns groups of channels that share name + host + type.
// GET /api/channel/duplicates
func ListDuplicateChannels(c *gin.Context) {
	groups, err := model.FindDuplicateChannelGroups()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if groups == nil {
		groups = []model.DuplicateChannelGroup{}
	}
	common.ApiSuccess(c, gin.H{"groups": groups})
}

// PreviewChannelMerge validates a merge without writing.
// POST /api/channel/merge/preview
func PreviewChannelMerge(c *gin.Context) {
	var req channelMergeBody
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}
	writeMergePreview(c, req.Ids, req.PrimaryId)
}

// MergeChannels merges duplicate channels into one multi-key channel.
// POST /api/channel/merge
// Preview-only calls should use POST /api/channel/merge/preview.
func MergeChannels(c *gin.Context) {
	var req channelMergeBody
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "invalid request")
		return
	}

	result, err := model.MergeChannels(req.Ids, req.PrimaryId)
	if err != nil {
		common.ApiErrorMsg(c, mergeErrorMessage(err))
		return
	}

	service.AfterChannelMutation()
	recordManageAudit(c, "channel.merge", map[string]interface{}{
		"primary_id":       result.PrimaryId,
		"merged_key_count": result.MergedKeyCount,
		"deleted_ids":      result.DeletedIds,
		"models_count":     result.ModelsCount,
	})

	common.ApiSuccess(c, result)
}

func writeMergePreview(c *gin.Context, ids []int, primaryId int) {
	preview, err := model.PreviewChannelMerge(ids, primaryId)
	if err != nil {
		common.ApiErrorMsg(c, mergeErrorMessage(err))
		return
	}
	common.ApiSuccess(c, preview)
}

func mergeErrorMessage(err error) string {
	switch {
	case errors.Is(err, model.ErrChannelMergeTooFew):
		return "至少需要选择 2 条渠道才能合并"
	case errors.Is(err, model.ErrChannelMergeNotFound):
		return "部分渠道不存在或已被删除"
	case errors.Is(err, model.ErrChannelMergeMismatch):
		return "只能合并同名、同 host、同类型的渠道"
	case errors.Is(err, model.ErrChannelMergeEmptyHost):
		return "渠道缺少可解析的 base_url host，无法合并"
	case errors.Is(err, model.ErrChannelMergePrimary):
		return "primary_id 必须是待合并渠道之一"
	case errors.Is(err, model.ErrChannelMergeNoKeys):
		return "合并后没有可用 key"
	default:
		if err == nil {
			return "merge failed"
		}
		return err.Error()
	}
}
