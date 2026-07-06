package handles

import (
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

func ListAuditLogs(c *gin.Context) {
	var req model.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	filter := &db.AuditLogFilter{
		Username: c.Query("username"),
		Action:   c.Query("action"),
		Path:     c.Query("path"),
	}
	if s := c.Query("start"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		filter.Start = t
	}
	if s := c.Query("end"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		filter.End = t
	}
	logs, total, err := db.GetAuditLogs(req.Page, req.PerPage, filter)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: logs,
		Total:   total,
	})
}

type ClearAuditLogsReq struct {
	Before string `json:"before"` // RFC3339; empty means delete all
}

func ClearAuditLogs(c *gin.Context) {
	var req ClearAuditLogsReq
	_ = c.ShouldBind(&req) // empty body is fine
	if req.Before != "" {
		t, err := time.Parse(time.RFC3339, req.Before)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		if err := db.DeleteAuditLogsBefore(t); err != nil {
			common.ErrorResp(c, err, 500, true)
			return
		}
	} else {
		if err := db.ClearAuditLogs(); err != nil {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	common.SuccessResp(c)
}
