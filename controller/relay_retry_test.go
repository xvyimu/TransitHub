package controller

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/xvyimu/TransitHub/dto"
	"github.com/xvyimu/TransitHub/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestShouldRetryStopsAfterChannelSelectionFailure(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	err := types.NewError(errors.New("no eligible channel"), types.ErrorCodeGetChannelFailed, types.ErrOptionWithSkipRetry())
	require.False(t, shouldRetry(ctx, err, 2))
}

func TestShouldRetrySwitchesChannelOnUpstreamQuotaExhaustion(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	err := types.WithOpenAIError(types.OpenAIError{
		Message: "upstream account has insufficient balance",
		Code:    "insufficient_quota",
	}, http.StatusTooManyRequests)
	require.True(t, shouldRetry(ctx, err, 2))
}

func TestShouldRetryDoesNotSwitchForLocalUserQuota(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	err := types.NewErrorWithStatusCode(
		errors.New("user quota insufficient"),
		types.ErrorCodeInsufficientUserQuota,
		http.StatusForbidden,
		types.ErrOptionWithSkipRetry(),
	)
	require.False(t, shouldRetry(ctx, err, 2))
}

func TestShouldRetryStopsWhenRequestContextCanceled(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.local/v1/chat/completions", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	ginCtx, _ := gin.CreateTestContext(nil)
	ginCtx.Request = req

	apiErr := types.NewError(errors.New("upstream 502"), types.ErrorCodeBadResponseStatusCode)
	apiErr.StatusCode = http.StatusBadGateway
	require.False(t, shouldRetry(ginCtx, apiErr, 2))
}

func TestShouldRetryTaskRelayStopsWhenRequestContextCanceled(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://example.local/suno/submit/music", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	ginCtx, _ := gin.CreateTestContext(nil)
	ginCtx.Request = req

	taskErr := &dto.TaskError{StatusCode: http.StatusBadGateway, LocalError: false}
	require.False(t, shouldRetryTaskRelay(ginCtx, 1, taskErr, 2))
}
