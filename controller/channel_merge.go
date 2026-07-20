package controller

import (
	"errors"
	"net/http"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type channelMergeBody struct {
	Ids       []int `json:"ids"`
	PrimaryId int   `json:"primary_id"`
	DryRun    bool  `json:"dry_run"`
}

// ListDuplicateChannels returns groups of channels that share name + host + type.
// GET /api/channel/duplicates
func ListDuplicateChannels(c *gin.Context) {
	groups, err := model.FindDuplicateChannelGroups()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if groups == nil {
		groups = []model.DuplicateChannelGroup{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"groups": groups,
		},
	})
}

// PreviewChannelMerge validates a merge without writing.
// POST /api/channel/merge/preview
func PreviewChannelMerge(c *gin.Context) {
	var req channelMergeBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	preview, err := model.PreviewChannelMerge(req.Ids, req.PrimaryId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": mergeErrorMessage(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    preview,
	})
}

// MergeChannels merges duplicate channels into one multi-key channel.
// POST /api/channel/merge
// When dry_run=true, returns the same payload as preview without writing.
func MergeChannels(c *gin.Context) {
	var req channelMergeBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "invalid request"})
		return
	}
	if req.DryRun {
		preview, err := model.PreviewChannelMerge(req.Ids, req.PrimaryId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": mergeErrorMessage(err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
			"data":    preview,
		})
		return
	}

	result, err := model.MergeChannels(req.Ids, req.PrimaryId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": mergeErrorMessage(err)})
		return
	}

	model.InitChannelCache()
	service.ResetProxyClientCache()
	recordManageAudit(c, "channel.merge", map[string]interface{}{
		"primary_id":       result.PrimaryId,
		"merged_key_count": result.MergedKeyCount,
		"deleted_ids":      result.DeletedIds,
		"models_count":     result.ModelsCount,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    result,
	})
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
