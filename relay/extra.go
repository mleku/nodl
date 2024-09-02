package relay

import (
	"context"

	. "nostr.mleku.dev"
)

const AuthContextKey = iota

func GetAuthStatus(ctx context.Context) (pubkey B, ok bool) {
	authedPubkey := ctx.Value(AuthContextKey)
	if authedPubkey == nil {
		return
	}
	return authedPubkey.(B), true
}
