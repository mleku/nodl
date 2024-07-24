package del

import "bytes"

type Items []B

func (c Items) Len() int           { return len(c) }
func (c Items) Less(i, j int) bool { return bytes.Compare(c[i], c[j]) < 0 }
func (c Items) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
