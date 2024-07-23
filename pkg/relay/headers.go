package relay

import (
	"net/http"
	"sort"
)

type Headers [][]string

func (h Headers) Len() int { return len(h) }
func (h Headers) Less(i, j int) bool { return h[i][0] < h[j][0] }
func (h Headers) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func SprintHeader(hdr http.Header) func() (s string) {
	return func() (s string) {
		var sections Headers
		s += "\n"
		for i := range hdr {
			sect := []string{i}
			sect = append(sect, hdr[i]...)
			sections = append(sections, sect)
		}
		sort.Sort(sections)
		for i := range sections {
			for j := range sections[i] {
				s += "\"" + sections[i][j] + "\" "
			}
			s += "\n"
		}
		return
	}
}

