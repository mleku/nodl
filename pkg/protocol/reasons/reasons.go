package reasons

import "bytes"

type Reason B

func (r Reason) S() S                   { return S(r) }
func (r Reason) B() B                   { return B(r) }
func (r Reason) IsPrefix(reason B) bool { return bytes.HasPrefix(reason, r.B()) }

var (
	AuthRequired = Reason("auth-required")
	PoW          = Reason("pow")
	Duplicate    = Reason("duplicate")
	Blocked      = Reason("blocked")
	RateLimited  = Reason("rate-limited")
	Invalid      = Reason("invalid")
	Error        = Reason("error")
	Unsupported  = Reason("unsupported")
)
