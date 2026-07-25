package setting

// Waffo Pancake hosted checkout configuration. Gateway is enabled once
// MerchantID + PrivateKey + ProductID are populated (no separate Enabled
// flag, matching Stripe / Creem). StoreID + ProductID are operator-bound
// via SaveWaffoPancakeConfig.
var (
	WaffoPancakeMerchantID string
	WaffoPancakePrivateKey string
	WaffoPancakeReturnURL  string
	WaffoPancakeUnitPrice  float64 = 1.0
	WaffoPancakeMinTopUp   int     = 1
	// WaffoPancakeMaxTopUp bounds a single Pancake top-up (in dollars) at
	// request validation time, mirroring the epay/stripe/waffo upper bound so
	// an oversized amount cannot reach quota conversion.
	WaffoPancakeMaxTopUp  int = 4000
	WaffoPancakeStoreID   string
	WaffoPancakeProductID string
)
