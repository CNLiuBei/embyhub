package task

import (
	"fmt"
	"log"
	"time"

	"embyhub/internal/service"
)

// SyncTask 同步任务
type SyncTask struct {
	embyService *service.EmbyService
	interval    time.Duration
	stopChan    chan struct{}
}

// NewSyncTask 创建同步任务
func NewSyncTask(interval time.Duration) *SyncTask {
	return &SyncTask{
		embyService: service.NewEmbyService(),
		interval:    interval,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动后台同步任务
func (t *SyncTask) Start() {
	log.Printf("[SyncTask] 后台同步任务已启动，间隔: %v", t.interval)

	// 启动时立即执行一次同步
	go t.runSync()

	// 定时执行
	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.runSync()
			case <-t.stopChan:
				ticker.Stop()
				log.Println("[SyncTask] 后台同步任务已停止")
				return
			}
		}
	}()
}

// Stop 停止同步任务
func (t *SyncTask) Stop() {
	close(t.stopChan)
}

// runSync 执行同步
func (t *SyncTask) runSync() {
	log.Println("[SyncTask] 开始执行Emby用户同步...")

	count, err := t.embyService.SyncUsers()
	if err != nil {
		log.Printf("[SyncTask] Emby用户同步失败: %v", err)
		return
	}

	if count > 0 {
		log.Printf("[SyncTask] Emby用户同步完成，新增/更新 %d 个用户", count)
	} else {
		fmt.Println("[SyncTask] Emby用户同步完成，无新用户")
	}
}
