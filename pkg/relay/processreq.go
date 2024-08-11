package relay

import (
	"errors"
	"sync"

	"git.replicatr.dev/pkg/codec/envelopes/closedenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/eoseenvelope"
	"git.replicatr.dev/pkg/codec/envelopes/reqenvelope"
	"git.replicatr.dev/pkg/protocol/reasons"
	"git.replicatr.dev/pkg/util/context"
)

func (rl *R) processReqEnvelope(msg B, env *reqenvelope.T, c Ctx, ws WS, svcURL S) (err error) {
	if !rl.IsAuthed(c, reqenvelope.L) {
		log.W.Ln("not authed")
		return
	}
	wg := sync.WaitGroup{}
	// a context just for the "stored events" request handler
	reqCtx, cancelReqCtx := context.CancelCause(c)
	// expose subscription id in the context
	reqCtx = context.Value(reqCtx, subscriptionIdKey, env.Subscription)
	// handle each filter separately -- dispatching events as they're loaded
	// from databases
	for _, f := range env.Filters.F {
		if err = rl.handleFilter(handleFilterParams{reqCtx, env.Subscription, &wg, ws, f}); log.T.Chk(err) {
			// fail everything if any filter is rejected
			reason := B(err.Error())
			if reasons.AuthRequired.IsPrefix(reason) {
				RequestAuth(c, env.Label())
			}
			if reasons.Blocked.IsPrefix(reason) {
				return
			}
			chk.E(closedenvelope.NewFrom(env.Subscription, reason).Write(ws))
			log.I.Ln("cancelling req context")
			cancelReqCtx(errors.New("filter rejected"))
			return
		}
	}
	go func() {
		// when all events have been loaded from databases and dispatched
		// we can cancel the context and fire the EOSE message
		wg.Wait()
		// log.I.Ln("cancelling req context")
		cancelReqCtx(nil)
		chk.E(eoseenvelope.NewFrom(env.Subscription).Write(ws))
	}()
	SetListener(env.Subscription.String(), ws, env.Filters, cancelReqCtx)
	return
}
