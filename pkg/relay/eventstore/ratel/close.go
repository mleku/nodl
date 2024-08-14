package ratel

func (r *T) Close() (err E) {
	log.T.F("closing database %s", r.Path)
	if err = r.DB.Flatten(4); chk.E(err) {
		return
	}
	if err = r.DB.Close(); chk.E(err) {
		return
	}
	if err = r.seq.Release(); chk.E(err) {
		return
	}
	return
}
