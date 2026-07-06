package audit

import (
	"context"
	stdpath "path"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/go-cache"
	log "github.com/sirupsen/logrus"
)

const (
	chanSize      = 1024
	batchSize     = 100
	flushInterval = 2 * time.Second
	// repeated reads of the same file by the same user/ip within this window
	// are logged once (video players and retries fire many range requests)
	dedupeWindow = 5 * time.Minute
)

var (
	logCh       chan model.AuditLog
	done        chan struct{}
	dedupeCache = cache.NewMemCache[struct{}]()
)

func Enabled() bool {
	return setting.GetBool(conf.AuditEnabled)
}

// Record logs an audit entry, extracting user/via/ip from ctx. Contexts that
// carry neither a user nor a via marker are internal and are skipped.
func Record(ctx context.Context, action, path string, size int64, detail string) {
	user, _ := ctx.Value(conf.UserKey).(*model.User)
	via, _ := ctx.Value(conf.AuditViaKey).(string)
	if user == nil && via == "" {
		return
	}
	// the direct-link middleware distinguishes explicit downloads from previews
	if action == model.AuditActionDownload {
		if intent, ok := ctx.Value(conf.AuditIntentKey).(string); ok && intent != "" {
			action = intent
		}
	}
	entry := model.AuditLog{
		Via:    via,
		Action: action,
		Path:   path,
		Detail: detail,
		Size:   size,
	}
	if user != nil {
		entry.UserID = user.ID
		entry.Username = user.Username
	} else if username, ok := ctx.Value(conf.AuditUsernameKey).(string); ok {
		// username embedded in a verified download sign
		entry.Username = username
	}
	if ip, ok := ctx.Value(conf.ClientIPKey).(string); ok {
		entry.IP = ip
	}
	RecordEntry(entry)
}

// RecordEntry logs an audit entry as is (except CreatedAt), for callers that
// don't have a request context, e.g. task callbacks.
func RecordEntry(entry model.AuditLog) {
	if logCh == nil || !Enabled() || shouldSkip(&entry) {
		return
	}
	entry.CreatedAt = time.Now()
	select {
	case logCh <- entry:
	default:
		log.Warnf("audit log channel is full, dropping entry: %s %s", entry.Action, entry.Path)
	}
}

// shouldSkip drops read entries for ignored file names (frontend-rendered
// helper files like README.md) and deduplicates repeated reads.
func shouldSkip(entry *model.AuditLog) bool {
	if entry.Action != model.AuditActionDownload && entry.Action != model.AuditActionPreview {
		return false
	}
	base := strings.ToLower(stdpath.Base(entry.Path))
	for name := range strings.SplitSeq(setting.GetStr(conf.AuditIgnoreNames), "\n") {
		if name = strings.TrimSpace(name); name != "" && strings.ToLower(name) == base {
			return true
		}
	}
	key := entry.Username + "|" + entry.IP + "|" + entry.Action + "|" + entry.Path
	if _, ok := dedupeCache.Get(key); ok {
		return true
	}
	dedupeCache.Set(key, struct{}{}, cache.WithEx[struct{}](dedupeWindow))
	return false
}

func InitAudit() {
	logCh = make(chan model.AuditLog, chanSize)
	done = make(chan struct{})
	go worker()
}

// Close flushes pending entries and stops the worker.
func Close() {
	if logCh == nil {
		return
	}
	close(logCh)
	<-done
	logCh = nil
}

func worker() {
	defer close(done)
	batch := make([]model.AuditLog, 0, batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := db.CreateAuditLogs(batch); err != nil {
			log.Errorf("failed to save %d audit logs: %+v", len(batch), err)
		}
		batch = batch[:0]
	}
	cleanup()
	flushTicker := time.NewTicker(flushInterval)
	defer flushTicker.Stop()
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()
	for {
		select {
		case entry, ok := <-logCh:
			if !ok {
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) >= batchSize {
				flush()
			}
		case <-flushTicker.C:
			flush()
		case <-cleanupTicker.C:
			flush()
			cleanup()
		}
	}
}

func cleanup() {
	days := setting.GetInt(conf.AuditRetentionDays, 0)
	if days <= 0 {
		return
	}
	if err := db.DeleteAuditLogsBefore(time.Now().AddDate(0, 0, -days)); err != nil {
		log.Errorf("failed to clean up expired audit logs: %+v", err)
	}
}
