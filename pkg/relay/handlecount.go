package relay

import (
	"strings"

	"git.replicatr.dev/pkg/codec/filter"
)

func (rl *R) handleCountRequest(c Ctx, id SubID, ws WS, f *filter.T) (subtotal int) {
	log.T.Ln("running count method")
	// overwrite the filter (for example, to eliminate some kinds or tags that we
	// know we don't support)
	for _, ovw := range rl.OverwriteCountFilter {
		ovw(c, f)
	}
	// then check if we'll reject this filter
	for _, reject := range rl.RejectReqFilters {
		if rej, msg := reject(c, id, f); rej {
			chk.E(NewNotice(msg).Write(ws))
			return 0
		}
	}
	// run the functions to count (generally it will be just one)
	var err error
	var res int
	for _, count := range rl.CountEvents {
		if res, err = count(c, f); err != nil {
			if strings.HasSuffix(err.Error(), "No events found") {
				log.E.Ln(err.Error())
			}
			chk.E(NewNotice(B(err.Error())).Write(ws))
		}
		subtotal += res
	}
	return
}
