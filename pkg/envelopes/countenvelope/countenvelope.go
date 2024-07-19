package countenvelope

import "C"
import (
	"bytes"

	"github.com/mleku/nodl/pkg/envelopes"
	"github.com/mleku/nodl/pkg/filters"
	"github.com/mleku/nodl/pkg/ints"
	"github.com/mleku/nodl/pkg/subscriptionid"
	"github.com/mleku/nodl/pkg/text"
)

const L = "COUNT"

type Request struct {
	ID      *subscriptionid.T
	Filters *filters.T
}

func New() *Request {
	return &Request{ID: subscriptionid.NewStd(), Filters: filters.New()}
}

func NewRequest(id *subscriptionid.T, filters *filters.T) *Request {
	return &Request{ID: id, Filters: filters}
}

func (req *Request) Label() string { return L }

func (req *Request) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = req.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			if o, err = req.Filters.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func (req *Request) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	if req.ID, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if rem, err = req.ID.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	req.Filters = filters.New()
	if rem, err = req.Filters.UnmarshalJSON(rem); chk.E(err) {
		return
	}
	// expect close brackets here but actually doesn't matter if neither
	// previous blocks failed
	rem = rem[:0]
	return
}

type Response struct {
	ID          *subscriptionid.T
	Count       int
	Approximate bool
}

func (res *Response) Label() string { return L }

func (res *Response) Marshal(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = res.ID.MarshalJSON(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			c := ints.New(res.Count)
			o, err = c.MarshalJSON(o)
			if res.Approximate {
				o = append(dst, ',')
				o = append(o, "true"...)
			}
			return
		})
	return
}

func (res *Response) UnmarshalJSON(b B) (rem B, err error) {
	rem = b
	var inID, inCount bool
	for ; len(rem) > 0; rem = rem[1:] {
		// first we should be finding a subscription ID
		if !inID && rem[0] == '"' {
			rem = rem[1:]
			// so we don't do this twice
			inID = true
			for i := range rem {
				if rem[i] == '\\' {
					continue
				} else if rem[i] == '"' {
					// skip escaped quotes
					if i > 0 {
						if rem[i-1] != '\\' {
							continue
						}
					}
					if res.ID, err = subscriptionid.
						New(text.NostrUnescape(rem[:i])); chk.E(err) {

						return
					}
					// trim the rest
					rem = rem[i:]
				}
			}
		} else {
			// pass the comma
			if rem[0] == ',' {
				continue
			} else if !inCount {
				inCount = true
				n := ints.New(0)
				if rem, err = n.UnmarshalJSON(rem); chk.E(err) {
					return
				}
				res.Count = int(n.Uint64())
			} else {
				// can only be either the end or optional approx
				if rem[0] == ']' {
					return
				} else {
					for i := range rem {
						if rem[i] == ']' {
							if bytes.Contains(rem[:i], B("true")) {
								res.Approximate = true
							}
							return
						}
					}
				}
			}
		}
	}
	return
}
