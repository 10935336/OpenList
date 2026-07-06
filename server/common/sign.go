package common

import (
	stdpath "path"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/internal/sign"
)

// UserSignName returns the username to embed in download signs, empty for
// guest/anonymous contexts.
func UserSignName(user *model.User) string {
	if user == nil || user.IsGuest() {
		return ""
	}
	return user.Username
}

func Sign(user *model.User, obj model.Obj, parent string, encrypt bool) string {
	if obj.IsDir() || (!encrypt && !setting.GetBool(conf.SignAll)) {
		return ""
	}
	return sign.SignWithUser(UserSignName(user), stdpath.Join(parent, obj.GetName()))
}
