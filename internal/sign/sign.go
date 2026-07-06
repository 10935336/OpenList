package sign

import (
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/pkg/sign"
)

var once sync.Once
var instance sign.Sign

func Sign(data string) string {
	expire := setting.GetInt(conf.LinkExpiration, 0)
	if expire == 0 {
		return NotExpired(data)
	} else {
		return WithDuration(data, time.Duration(expire)*time.Hour)
	}
}

// SignWithUser signs data with the username embedded, so the user can be
// identified when the link is used. Format: "<hmac>:<expire>:<b64url(username)>",
// where the hmac covers "username\x00data". Verify accepts both formats.
func SignWithUser(username, data string) string {
	if username == "" {
		return Sign(data)
	}
	return Sign(username+"\x00"+data) + ":" + base64.URLEncoding.EncodeToString([]byte(username))
}

func WithDuration(data string, d time.Duration) string {
	once.Do(Instance)
	return instance.Sign(data, time.Now().Add(d).Unix())
}

func NotExpired(data string) string {
	once.Do(Instance)
	return instance.Sign(data, 0)
}

func Verify(data string, s string) error {
	once.Do(Instance)
	if username, rest, ok := splitUserSign(s); ok {
		return instance.Verify(username+"\x00"+data, rest)
	}
	return instance.Verify(data, s)
}

// UserFromSign extracts the embedded username from a user-signed string.
// It does NOT verify the signature; only use it after Verify succeeded.
func UserFromSign(s string) string {
	username, _, ok := splitUserSign(s)
	if !ok {
		return ""
	}
	return username
}

// splitUserSign splits "<hmac>:<expire>:<b64url(username)>" into the username
// and the plain "<hmac>:<expire>" part. Returns ok=false for legacy signs.
func splitUserSign(s string) (username, rest string, ok bool) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return "", "", false
	}
	name, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil || len(name) == 0 {
		return "", "", false
	}
	return string(name), parts[0] + ":" + parts[1], true
}

func Instance() {
	instance = sign.NewHMACSign([]byte(setting.GetStr(conf.Token)))
}
