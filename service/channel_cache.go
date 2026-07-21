package service

import "github.com/xvyimu/TransitHub/model"

// AfterChannelMutation refreshes channel routing cache and proxy HTTP clients
// after channel / ability mutations. Controllers should call this instead of
// pairing model.InitChannelCache + ResetProxyClientCache ad hoc (easy to miss
// one of the two and serve stale proxy settings).
func AfterChannelMutation() {
	model.InitChannelCache()
	ResetProxyClientCache()
}
