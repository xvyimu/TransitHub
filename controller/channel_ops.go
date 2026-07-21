package controller

import (
	"strconv"
	"strings"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/model"
	"github.com/xvyimu/TransitHub/service"

	"github.com/gin-gonic/gin"
)

func GetChannelOps(c *gin.Context) {
	common.ApiSuccess(c, gin.H{
		"retry_times": common.RetryTimes,
	})
}

// GetChannelHealthMetrics returns in-process relay/circuit/shadow/refund metrics (WP-D).
func GetChannelHealthMetrics(c *gin.Context) {
	common.ApiSuccess(c, service.SnapshotChannelHealth())
}

// ListRefundIntents returns recent refund outbox rows for reconciliation (WP-F).
func ListRefundIntents(c *gin.Context) {
	status := strings.TrimSpace(c.Query("status"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := model.ListRefundIntents(status, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	counts, _ := model.CountRefundIntentsByStatus()
	common.ApiSuccess(c, gin.H{
		"items":  items,
		"counts": counts,
	})
}
