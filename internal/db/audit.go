package db

import (
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/pkg/errors"
)

type AuditLogFilter struct {
	Username string
	Action   string
	Path     string
	Start    time.Time
	End      time.Time
}

func CreateAuditLogs(logs []model.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return errors.WithStack(db.Create(&logs).Error)
}

func GetAuditLogs(pageIndex, pageSize int, filter *AuditLogFilter) (logs []model.AuditLog, count int64, err error) {
	auditDB := db.Model(&model.AuditLog{})
	if filter != nil {
		if filter.Username != "" {
			auditDB = auditDB.Where(columnName("username")+" LIKE ?", "%"+filter.Username+"%")
		}
		if filter.Action != "" {
			auditDB = auditDB.Where(columnName("action")+" = ?", filter.Action)
		}
		if filter.Path != "" {
			auditDB = auditDB.Where(columnName("path")+" LIKE ?", "%"+filter.Path+"%")
		}
		if !filter.Start.IsZero() {
			auditDB = auditDB.Where(columnName("created_at")+" >= ?", filter.Start)
		}
		if !filter.End.IsZero() {
			auditDB = auditDB.Where(columnName("created_at")+" < ?", filter.End)
		}
	}
	if err = auditDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get audit logs count")
	}
	if err = auditDB.Order(columnName("id") + " DESC").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get find audit logs")
	}
	return logs, count, nil
}

func DeleteAuditLogsBefore(t time.Time) error {
	return errors.WithStack(db.Where(columnName("created_at")+" < ?", t).Delete(&model.AuditLog{}).Error)
}

func ClearAuditLogs() error {
	return errors.WithStack(db.Where(columnName("id") + " > 0").Delete(&model.AuditLog{}).Error)
}
