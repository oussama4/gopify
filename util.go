package gopify

import (
	"crypto/rand"
	"strings"
)

func uniqueToken(size int) string {
	alphabet := []byte("ModuleSymbhasOwnPr-0123456789ABCDEFGHNRVfgctiUvz_KqYTJkLxpZXIjQW")
	bytes := make([]byte, size)
	rand.Read(bytes)

	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteByte(alphabet[bytes[i]&63])
	}

	return b.String()
}
