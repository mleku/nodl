package main

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"strings"
)

//go:embed util.go
var common []byte

func main() {
	filepath.Walk(".", func(path S, info os.FileInfo, err error) error {
		if path != "." && info.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return filepath.SkipDir
			}
			log.I.Ln(path)
			log.I.Ln(filepath.Base(path))
			b := common
			if filepath.Base(path) == "cmd" ||
				filepath.Base(path) == "pkg" ||
				filepath.Base(path) == "internal" ||
				strings.HasPrefix(filepath.Base(path), "gen") ||
				filepath.Base(path) == "common" ||
				filepath.Base(path) == "lol" ||
				filepath.Base(path) == "precomps" ||
				filepath.Base(path) == "atomic" {
				return nil
			}
			if !strings.HasPrefix(path, "cmd"+S(filepath.Separator)) {
				b = bytes.Replace(
					common,
					B("package main"),
					B("package "+filepath.Base(path)),
					1,
				)
			}
			_ = b
			log.I.F("\n%s", S(b))
			chk.E(os.WriteFile(filepath.Join(path, "util.go"), b, 0660))
		}
		return nil
	})
}
