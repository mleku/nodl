package relay

import (
	"github.com/mleku/nodl/pkg/codec/filter"
)

var MaxLimit int

func FilterClampLimit(c Ctx, f *filter.T) {
	if f.Limit < 1 || f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}
