package stat

import (
	"os"
)

func IsFresh(a, b os.FileInfo) bool {
	sizeOk := a.Size() == b.Size()
	modeOk := a.Mode() == b.Mode()
	modTimeOk := a.ModTime() == b.ModTime()
	return sizeOk && modeOk && modTimeOk
}
