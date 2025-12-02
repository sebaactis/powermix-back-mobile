package coffeeji

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

func buildAuthHeaders(key, secret string) (timestamp string, keyMd5 string) {
	t := time.Now().UnixMilli()
	timestamp = fmt.Sprintf("%d", t)

	raw := key + secret + timestamp

	sum := md5.Sum([]byte(raw))
	keyMd5 = hex.EncodeToString(sum[:])

	return timestamp, keyMd5
}
