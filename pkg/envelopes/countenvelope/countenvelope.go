package countenvelope

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
	ID      subscriptionid.T
	Filters filters.T
}

func (req *Request) Label() string { return L }

func (req *Request) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = req.ID.Marshal(o); chk.E(err) {
				return
			}
			dst = append(dst, ',')
			if o, err = req.Filters.MarshalJSON(o); chk.E(err) {
				return
			}
			return
		})
	return
}

func UnmarshalRequest(b B) (req *Request, rem B, err error) {
	rem = b
	req = &Request{}
	var inID bool
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
					req.ID = text.NostrUnescape(rem[:i])
					// trim the rest
					rem = rem[i:]
				}
			}
		} else {
			// second should be filters
			var fa any
			if fa, rem, err = filters.New().UnmarshalJSON(rem); chk.E(err) {
				return
			}
			req.Filters = fa.(filters.T)
			// literally can't be anything more after this
			return
		}
	}
	return
}

type Response struct {
	ID          subscriptionid.T
	Count       int
	Approximate bool
}

func (res *Response) Label() string { return L }

func (res *Response) Marshal(dst B) (b B, err error) {
	b = dst
	b, err = envelopes.Marshal(b, L,
		func(bst B) (o B, err error) {
			o = bst
			if o, err = res.ID.Marshal(o); chk.E(err) {
				return
			}
			o = append(o, ',')
			o, err = ints.T(res.Count).MarshalJSON(o)
			if res.Approximate {
				o = append(dst, ',')
				o = append(o, "true"...)
			}
			return
		})
	return
}

func UnmarshalResponse(b B) (res *Response, rem B, err error) {
	rem = b
	res = &Response{}
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
					res.ID = text.NostrUnescape(rem[:i])
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
				var n any
				if n, rem, err = ints.New().UnmarshalJSON(rem); chk.E(err) {
					return
				}
				res.Count = int(n.(ints.T))
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
