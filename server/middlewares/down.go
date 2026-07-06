package middlewares

import (
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"

	"github.com/OpenListTeam/OpenList/v4/internal/errs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/sign"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func PathParse(c *gin.Context) {
	rawPath := parsePath(c.Param("path"))
	common.GinAppendValues(c, conf.PathKey, rawPath)
	c.Next()
}

func Down(verifyFunc func(string, string) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		rawPath := c.Request.Context().Value(conf.PathKey).(string)
		meta, err := op.GetNearestMeta(rawPath)
		if err != nil && !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorPage(c, err, 500, true)
			return
		}
		// distinguish explicit downloads (dl=1, added by the frontend download
		// buttons) from previews/streaming for the audit log
		intent := model.AuditActionPreview
		if _, ok := c.GetQuery("dl"); ok {
			intent = model.AuditActionDownload
		}
		common.GinAppendValues(c, conf.MetaKey, meta,
			conf.AuditViaKey, "direct", conf.ClientIPKey, c.ClientIP(),
			conf.AuditIntentKey, intent)
		// verify sign
		if needSign(meta, rawPath) {
			s := strings.TrimSuffix(c.Query("sign"), "/")
			err = verifyFunc(rawPath, s)
			if err != nil {
				common.ErrorPage(c, err, 401)
				c.Abort()
				return
			}
			// the sign is valid, so an embedded username is trustworthy
			if username := sign.UserFromSign(s); username != "" {
				common.GinAppendValues(c, conf.AuditUsernameKey, username)
			}
		}
		c.Next()
	}
}

// TODO: implement
// path maybe contains # ? etc.
func parsePath(path string) string {
	return utils.FixAndCleanPath(path)
}

func needSign(meta *model.Meta, path string) bool {
	if setting.GetBool(conf.SignAll) {
		return true
	}
	if common.IsStorageSignEnabled(path) {
		return true
	}
	if meta == nil || meta.Password == "" {
		return false
	}
	if !meta.PSub && path != meta.Path {
		return false
	}
	return true
}
