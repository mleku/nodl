package relay

import (
	"git.replicatr.dev/pkg/codec/filter"
	// 	"golang.org/x/exp/slices"
)

// NoComplexFilters disallows filters with more than 3 tags or total of 6 of kinds and tags in sum..
func NoComplexFilters(c Ctx, id SubID, f *filter.T) (reject bool, msg B) {
	items := f.Tags.Len() + f.Kinds.Len()
	if items > 6 && f.Tags.Len() > 3 {
		return true, B("too many things to filter for")
	}
	return
}

// NoEmptyFilters disallows filters that don't have at least a tag, a kind, an author or an id, or since or until.
func NoEmptyFilters(c Ctx, id SubID, f *filter.T) (reject bool, msg B) {
	cf := f.Kinds.Len() + f.IDs.Len() + f.Authors.Len()
	for _, tagItems := range f.Tags.T {
		cf += len(tagItems.Field)
	}
	if f.Since != nil {
		cf++
	}
	if f.Until != nil {
		cf++
	}
	if f.Limit > 0 {
		cf++
	}
	if cf == 0 {
		return true, B("can't handle empty filters")
	}
	return
}

// // AntiSyncBots tries to prevent people from syncing kind:1s from this relay to
// // else by always requiring an author parameter at least.
// func AntiSyncBots(c context.T, f *filter.T) (rej bool, msg string) {
// 	return (len(f.Kinds) == 0 ||
// 			slices.Contains(f.Kinds, 1)) &&
// 			len(f.Authors) == 0,
// 		"an author must be specified to get their kind:1 notes"
// }

func NoSearchQueries(c Ctx, id SubID, f *filter.T) (reject bool, msg B) {
	if len(f.Search) != 0 {
		return true, B("search is not supported")
	}
	return
}

// func RemoveSearchQueries(c context.T, f *filter.T) {
// 	f.Search = ""
// }
//
// func RemoveAllButKinds(k ...kind.T) OverwriteFilter {
// 	return func(c context.T, f *filter.T) {
// 		if n := len(f.Kinds); n > 0 {
// 			newKinds := make(kinds.T, 0, n)
// 			for i := 0; i < n; i++ {
// 				if kk := f.Kinds[i]; slices.Contains(k, kk) {
// 					newKinds = append(newKinds, kk)
// 				}
// 			}
// 			f.Kinds = newKinds
// 		}
// 	}
// }
//
// func LimitAuthorsAndIDs(authors, ids int) OverwriteFilter {
// 	return func(c context.T, f *filter.T) {
// 		if len(f.Authors) > authors {
// 			log.I.Ln("limiting authors to", authors)
// 			f.Authors = f.Authors[:20]
// 		}
// 		if len(f.IDs) > ids {
// 			log.I.Ln("limiting IDs to", ids)
// 			f.IDs = f.IDs[:20]
// 		}
// 	}
// }
//
// func RemoveAllButTags(tagNames ...string) OverwriteFilter {
// 	return func(c context.T, f *filter.T) {
// 		for tagName := range f.Tags {
// 			if !slices.Contains(tagNames, tagName) {
// 				delete(f.Tags, tagName)
// 			}
// 		}
// 	}
// }
