package filters

import (
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/filter"
)

type T struct {
	F []*filter.T
}

func New() (f *T) { return &T{} }

func (f *T) Match(event *event.T) bool {
	for _, f := range f.F {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

func (f *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	end := len(f.F) - 1
	for i := range f.F {
		if b, err = f.F[i].MarshalJSON(b); chk.E(err) {
			return
		}
		if i < end {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (f *T) UnmarshalJSON(b B) (rem B, err error) {
	rem = b[:]
	for len(rem) > 0 {
		switch rem[0] {
		case '[':
			if len(rem) > 1 && rem[1] == ']' {
				rem = rem[1:]
				return
			}
			ffa := filter.New()
			if rem, err = ffa.UnmarshalJSON(rem); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
			// continue
		case ',':
			rem = rem[1:]
			// next
		case ']':
			rem = rem[1:]
			// the end
			return
		}
	}
	return
}
