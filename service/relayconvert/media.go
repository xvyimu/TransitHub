package relayconvert

import relaymedia "github.com/xvyimu/TransitHub/service/relayconvert/internal/media"

type MediaResolver = relaymedia.MediaResolver

func SetMediaResolver(resolver MediaResolver) {
	relaymedia.SetMediaResolver(resolver)
}
