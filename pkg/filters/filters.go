package filters

import (
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/filter"
)

type T []*filter.T

func New() (f T) { return }

func (ff T) Match(event *event.T) bool {
	for _, f := range ff {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

func (ff T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	end := len(ff) - 1
	for i := range ff {
		if b, err = ff[i].MarshalJSON(b); chk.E(err) {
			return
		}
		if i < end {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (ff T) UnmarshalJSON(b B) (fa any, rem B, err error) {
	rem = b[:]
	for len(rem) > 0 {
		switch rem[0] {
		case '[':
			if len(rem) > 1 && rem[1] == ']' {
				rem = rem[1:]
				return
			}
			var ffa any
			if ffa, rem, err = filter.New().UnmarshalJSON(rem); chk.E(err) {
				return
			}
			ff = append(ff, ffa.(*filter.T))
			// continue
		case ',':
			rem = rem[1:]
			// next
		case ']':
			rem = rem[1:]
			// the end
			fa = ff
			return
		}
	}
	return
}
