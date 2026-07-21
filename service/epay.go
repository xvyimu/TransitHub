package service

import (
	"github.com/xvyimu/TransitHub/setting/operation_setting"
	"github.com/xvyimu/TransitHub/setting/system_setting"
)

func GetCallbackAddress() string {
	if operation_setting.CustomCallbackAddress == "" {
		return system_setting.ServerAddress
	}
	return operation_setting.CustomCallbackAddress
}
