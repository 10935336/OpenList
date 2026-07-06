package model

import "time"

// audit actions
const (
	AuditActionDownload = "download"
	AuditActionUpload   = "upload"
	AuditActionMkdir    = "mkdir"
	AuditActionRename   = "rename"
	AuditActionMove     = "move"
	AuditActionCopy     = "copy"
	AuditActionMerge    = "merge"
	AuditActionRemove   = "remove"
)

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username" gorm:"index;size:255"` // stored redundantly so records survive user deletion
	IP        string    `json:"ip" gorm:"size:64"`
	Via       string    `json:"via" gorm:"size:16"`          // web/direct/webdav/ftp/sftp/s3/mcp
	Action    string    `json:"action" gorm:"index;size:16"` // download/upload/mkdir/rename/move/copy/remove
	Path      string    `json:"path" gorm:"size:750"`
	Detail    string    `json:"detail" gorm:"size:750"` // e.g. new name for rename, destination dir for move/copy
	Size      int64     `json:"size"`
}
