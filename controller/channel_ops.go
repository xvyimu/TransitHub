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

// GetChannelHealthMetrics 返回进程内中继/熔断/shadow/退款指标（WP-D）。
// 约束：路由层须 AdminAuth + channel:operate；禁止匿名；计数重启清零（db_* 除外）。
func GetChannelHealthMetrics(c *gin.Context) {
	common.ApiSuccess(c, service.SnapshotChannelHealth())
}

// ListRefundIntents 管理端退款 outbox 对账列表（WP-F）。
// 约束：AdminAuth + channel:operate；响应含 items+counts；不得含 token key。
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
