package countenvelope

// func TestMarshalJSONUnmarshalJSON(t *testing.T) {
// 	var err error
// 	rb, rbc, rb2 := make(B, 0, 65535), make(B, 0, 65535), make(B, 0, 65535)
// 	// for _ = range 50 {
// 	var f *filters.T
// 	if f, err = filters.GenFilters(5); chk.E(err) {
// 		t.Fatal(err)
// 	}
// 	var s *subscriptionid.T
// 	if s = subscriptionid.NewStd(); chk.E(err) {
// 		t.Fatal(err)
// 	}
// 	req := NewRequest(s, f)
// 	if rb, err = req.MarshalJSON(rb); chk.E(err) {
// 		t.Fatal(err)
// 	}
// 	rbc = rbc[:len(rb)]
// 	copy(rbc, rb)
// 	// log.I.F("\n%d %s\n\n%d %s\n", len(rb), rb, len(rbc), rbc)
// 	req2 := New()
// 	var rem B
// 	if rem, err = req2.UnmarshalJSON(rb); chk.E(err) {
// 		t.Fatal(err)
// 	}
// 	if len(rem) > 0 {
// 		t.Fatalf("unmarshal failed, remainder %d %s", len(rem), rem)
// 	}
// 	// log.I.S(req2)
// 	if rb2, err = req2.MarshalJSON(rb2); chk.E(err) {
// 		t.Fatal(err)
// 	}
// 	if !equals(rbc, rb2) {
// 		t.Fatalf("unmarshal failed\n%d %s\n%d %s\n", len(rbc), rbc, len(rb2),
// 			rb2)
// 	}
// 	rb, rbc, rb2 = rb[:0], rbc[:0], rb2[:0]
// 	// }
// }
