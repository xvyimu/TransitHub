package service

import (
	"strings"

	"github.com/xvyimu/TransitHub/common"
	"github.com/xvyimu/TransitHub/setting/system_setting"
)

func PaymentReturnURL(suffix string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	return base + common.ThemeAwarePath(suffix)
}
