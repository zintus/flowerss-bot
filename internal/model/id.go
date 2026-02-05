package model

import (
	"encoding/hex"
	"hash/fnv"
)

func GenHashID(sLink string, id string, rawLink string) string {
	// When id (GUID) is empty, use rawLink as fallback to ensure uniqueness
	effectiveID := id
	if effectiveID == "" {
		effectiveID = rawLink
	}
	idString := sLink + "||" + effectiveID
	f := fnv.New64()
	f.Write([]byte(idString))

	encoded := hex.EncodeToString(f.Sum(nil))
	return encoded
}
