package email

import (
	"context"
	"sync"
	"time"
)

// QueueItem 队列项
type QueueItem struct {
	Message   *Message
	Retries   int
	CreatedAt time.Time
	Priority  int
}

// Queue 邮件队列
type Queue struct {
	service    *Service
	items      chan *QueueItem
	workers    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	maxRetries int
	retryDelay time.Duration
	metrics    *QueueMetrics
}

// QueueMetrics 队列指标
type QueueMetrics struct {
	TotalEnqueued int64
	TotalSent     int64
	TotalFailed   int64
	CurrentQueue  int64
	mu            sync.RWMutex
}

// NewQueue 创建邮件队列
func NewQueue(service *Service, workers int, queueSize int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	q := &Queue{
		service:    service,
		items:      make(chan *QueueItem, queueSize),
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
		maxRetries: 3,
		retryDelay: time.Second * 2,
		metrics:    &QueueMetrics{},
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	service.logger.Info("Email queue started", "workers", workers, "queue_size", queueSize)
	return q
}

// worker 工作协程
func (q *Queue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			q.service.logger.Info("Email queue worker stopping", "worker_id", id)
			return
		case item := <-q.items:
			q.processItem(item, id)
		}
	}
}

// processItem 处理队列项
func (q *Queue) processItem(item *QueueItem, workerID int) {
	q.metrics.mu.Lock()
	q.metrics.CurrentQueue--
	q.metrics.mu.Unlock()

	q.service.logger.Info("Processing email",
		"worker_id", workerID,
		"to", item.Message.To,
		"subject", item.Message.Subject,
		"retry", item.Retries,
	)

	err := q.service.Send(item.Message)

	if err != nil {
		q.service.logger.Error("Failed to send email from queue",
			"worker_id", workerID,
			"to", item.Message.To,
			"error", err,
			"retry", item.Retries,
		)

		// 重试逻辑
		if item.Retries < q.maxRetries {
			item.Retries++
			time.Sleep(q.retryDelay * time.Duration(item.Retries))

			select {
			case q.items <- item:
				q.metrics.mu.Lock()
				q.metrics.CurrentQueue++
				q.metrics.mu.Unlock()
				q.service.logger.Info("Email re-queued for retry",
					"worker_id", workerID,
					"to", item.Message.To,
					"retry", item.Retries,
				)
			default:
				q.service.logger.Error("Failed to re-queue email, queue full",
					"worker_id", workerID,
					"to", item.Message.To,
				)
				q.metrics.mu.Lock()
				q.metrics.TotalFailed++
				q.metrics.mu.Unlock()
			}
		} else {
			q.service.logger.Error("Email failed after max retries",
				"worker_id", workerID,
				"to", item.Message.To,
				"retries", item.Retries,
			)
			q.metrics.mu.Lock()
			q.metrics.TotalFailed++
			q.metrics.mu.Unlock()
		}
	} else {
		q.service.logger.Info("Email sent successfully",
			"worker_id", workerID,
			"to", item.Message.To,
		)
		q.metrics.mu.Lock()
		q.metrics.TotalSent++
		q.metrics.mu.Unlock()
	}
}

// Enqueue 添加到队列
func (q *Queue) Enqueue(msg *Message) error {
	item := &QueueItem{
		Message:   msg,
		Retries:   0,
		CreatedAt: time.Now(),
		Priority:  msg.Priority,
	}

	select {
	case q.items <- item:
		q.metrics.mu.Lock()
		q.metrics.TotalEnqueued++
		q.metrics.CurrentQueue++
		q.metrics.mu.Unlock()
		return nil
	default:
		q.service.logger.Error("Email queue is full", "to", msg.To)
		return ErrSendFailed
	}
}

// EnqueueHTML 添加HTML邮件到队列
func (q *Queue) EnqueueHTML(to, subject, body string) error {
	return q.Enqueue(&Message{
		To:          []string{to},
		Subject:     subject,
		Body:        body,
		ContentType: "text/html; charset=UTF-8",
	})
}

// EnqueueTemplate 使用模板添加邮件到队列
func (q *Queue) EnqueueTemplate(to, subject string, templateType TemplateType, data interface{}) error {
	body, err := q.service.templateManager.Render(templateType, data)
	if err != nil {
		q.service.logger.Error("Failed to render template", "template", templateType, "error", err)
		return err
	}

	return q.EnqueueHTML(to, subject, body)
}

// Stop 停止队列
func (q *Queue) Stop(timeout time.Duration) error {
	q.service.logger.Info("Stopping email queue", "timeout", timeout)

	// 停止接收新邮件
	q.cancel()

	// 等待现有邮件处理完成（带超时）
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		q.service.logger.Info("Email queue stopped gracefully")
		return nil
	case <-time.After(timeout):
		q.service.logger.Warn("Email queue stop timeout, forcing shutdown")
		return nil
	}
}

// GetMetrics 获取队列指标
func (q *Queue) GetMetrics() QueueMetrics {
	q.metrics.mu.RLock()
	defer q.metrics.mu.RUnlock()

	return QueueMetrics{
		TotalEnqueued: q.metrics.TotalEnqueued,
		TotalSent:     q.metrics.TotalSent,
		TotalFailed:   q.metrics.TotalFailed,
		CurrentQueue:  q.metrics.CurrentQueue,
	}
}

// QueueSize 获取当前队列大小
func (q *Queue) QueueSize() int {
	return len(q.items)
}

// IsHealthy 检查队列健康状态
func (q *Queue) IsHealthy() bool {
	metrics := q.GetMetrics()

	// 如果失败率过高，认为不健康
	if metrics.TotalSent+metrics.TotalFailed > 0 {
		failureRate := float64(metrics.TotalFailed) / float64(metrics.TotalSent+metrics.TotalFailed)
		return failureRate < 0.1 // 失败率小于10%认为健康
	}

	return true
}
