package oauth

import (
	"net/http"
	"sync"
	"time"

	"github.com/xvyimu/TransitHub/common"
)

var (
	oauthTransportOnce sync.Once
	oauthTransport     *http.Transport
)

func newHTTPClient(timeout time.Duration) *http.Client {
	oauthTransportOnce.Do(func() {
		oauthTransport = common.NewOutboundHTTPTransport(http.ProxyFromEnvironment, nil)
	})
	return &http.Client{Transport: oauthTransport, Timeout: timeout}
}
