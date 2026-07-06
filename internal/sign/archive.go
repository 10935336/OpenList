package sign

import (
	"encoding/base64"
	"sync"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/pkg/sign"
)

var onceArchive sync.Once
var instanceArchive sign.Sign

func SignArchive(data string) string {
	expire := setting.GetInt(conf.LinkExpiration, 0)
	if expire == 0 {
		return NotExpiredArchive(data)
	} else {
		return WithDurationArchive(data, time.Duration(expire)*time.Hour)
	}
}

// SignArchiveWithUser embeds the username like SignWithUser, see there.
func SignArchiveWithUser(username, data string) string {
	if username == "" {
		return SignArchive(data)
	}
	return SignArchive(username+"\x00"+data) + ":" + base64.URLEncoding.EncodeToString([]byte(username))
}

func WithDurationArchive(data string, d time.Duration) string {
	onceArchive.Do(InstanceArchive)
	return instanceArchive.Sign(data, time.Now().Add(d).Unix())
}

func NotExpiredArchive(data string) string {
	onceArchive.Do(InstanceArchive)
	return instanceArchive.Sign(data, 0)
}

func VerifyArchive(data string, s string) error {
	onceArchive.Do(InstanceArchive)
	if username, rest, ok := splitUserSign(s); ok {
		return instanceArchive.Verify(username+"\x00"+data, rest)
	}
	return instanceArchive.Verify(data, s)
}

func InstanceArchive() {
	instanceArchive = sign.NewHMACSign([]byte(setting.GetStr(conf.Token) + "-archive"))
}
