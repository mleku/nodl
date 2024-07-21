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

var _ envelopes.I = (*Request)(nil)

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

func (req *Request) UnmarshalJSON(b B) (r B, err error) {
	r = b
	if req.ID, err = subscriptionid.New(B{0}); chk.E(err) {
		return
	}
	if r, err = req.ID.UnmarshalJSON(r); chk.E(err) {
		return
	}
	req.Filters = filters.New()
	if r, err = req.Filters.UnmarshalJSON(r); chk.E(err) {
		return
	}
	if r, err = envelopes.SkipToTheEnd(r); chk.E(err) {
		return
	}
	return
}

type Response struct {
	ID          *subscriptionid.T
	Count       int
	Approximate bool
}

var _ envelopes.I = (*Response)(nil)

func (res *Response) Label() string { return L }

func (res *Response) MarshalJSON(dst B) (b B, err error) {
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

func (res *Response) UnmarshalJSON(b B) (r B, err error) {
	r = b
	var inID, inCount bool
	for ; len(r) > 0; r = r[1:] {
		// first we should be finding a subscription ID
		if !inID && r[0] == '"' {
			r = r[1:]
			// so we don't do this twice
			inID = true
			for i := range r {
				if r[i] == '\\' {
					continue
				} else if r[i] == '"' {
					// skip escaped quotes
					if i > 0 {
						if r[i-1] != '\\' {
							continue
						}
					}
					if res.ID, err = subscriptionid.
						New(text.NostrUnescape(r[:i])); chk.E(err) {

						return
					}
					// trim the rest
					r = r[i:]
				}
			}
		} else {
			// pass the comma
			if r[0] == ',' {
				continue
			} else if !inCount {
				inCount = true
				n := ints.New(0)
				if r, err = n.UnmarshalJSON(r); chk.E(err) {
					return
				}
				res.Count = int(n.Uint64())
			} else {
				// can only be either the end or optional approx
				if r[0] == ']' {
					return
				} else {
					for i := range r {
						if r[i] == ']' {
							if bytes.Contains(r[:i], B("true")) {
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
