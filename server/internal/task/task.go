package task

import (
	"fmt"
	"sync"
	"time"

	"remote-file-manager/internal/log"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 任务待处理
	TaskStatusRunning   TaskStatus = "running"   // 任务运行中
	TaskStatusCompleted TaskStatus = "completed" // 任务已完成
	TaskStatusFailed    TaskStatus = "failed"    // 任务失败
)

// Task 任务结构体
type Task struct {
	ID           string     `json:"id"`           // 任务ID
	Type         string     `json:"type"`         // 任务类型：zip, download等
	ConnID       string     `json:"connId"`       // 连接ID
	Path         string     `json:"path"`         // 目标路径
	Status       TaskStatus `json:"status"`       // 任务状态
	Progress     int        `json:"progress"`     // 任务进度（0-100）
	FileName     string     `json:"fileName"`     // 生成的文件名
	FilePath     string     `json:"filePath"`     // 生成的文件完整路径
	ErrorMessage string     `json:"errorMessage"` // 错误信息
	CreatedAt    time.Time  `json:"createdAt"`    // 任务创建时间
	UpdatedAt    time.Time  `json:"updatedAt"`    // 任务更新时间
	Content      []byte     `json:"-"`            // 缓存的文件内容，不序列化到JSON
}

// TaskManager 任务管理器
type TaskManager struct {
	tasks      map[string]*Task
	mutex      sync.Mutex
	nextTaskID int
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks:      make(map[string]*Task),
		mutex:      sync.Mutex{},
		nextTaskID: 0,
	}
}

// CreateTask 创建新任务
func (m *TaskManager) CreateTask(taskType, connID, path string) *Task {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 生成唯一任务ID
	m.nextTaskID++
	taskID := fmt.Sprintf("task-%d", m.nextTaskID)

	// 创建任务
	task := &Task{
		ID:           taskID,
		Type:         taskType,
		ConnID:       connID,
		Path:         path,
		Status:       TaskStatusPending,
		Progress:     0,
		FileName:     "",
		ErrorMessage: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 添加到任务列表
	m.tasks[taskID] = task

	log.Info("Created task: %s, type: %s, path: %s", taskID, taskType, path)

	return task
}

// GetTask 获取任务
func (m *TaskManager) GetTask(taskID string) (*Task, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	return task, exists
}

// UpdateTaskProgress 更新任务进度
func (m *TaskManager) UpdateTaskProgress(taskID string, progress int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		log.Warn("Task not found: %s", taskID)
		return
	}

	// 更新进度
	task.Progress = progress
	task.UpdatedAt = time.Now()

	// 如果进度达到100%，标记任务为已完成
	if progress >= 100 {
		task.Status = TaskStatusCompleted
	}

	log.Info("Updated task progress: %s, progress: %d%%", taskID, progress)
}

// UpdateTaskStatus 更新任务状态
func (m *TaskManager) UpdateTaskStatus(taskID string, status TaskStatus) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		log.Warn("Task not found: %s", taskID)
		return
	}

	// 更新状态
	task.Status = status
	task.UpdatedAt = time.Now()

	log.Info("Updated task status: %s, status: %s", taskID, status)
}

// UpdateTaskFileName 更新任务生成的文件名
func (m *TaskManager) UpdateTaskFileName(taskID string, fileName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		log.Warn("Task not found: %s", taskID)
		return
	}

	// 更新文件名
	task.FileName = fileName
	task.UpdatedAt = time.Now()

	log.Info("Updated task file name: %s, fileName: %s", taskID, fileName)
}

// UpdateTaskFilePath 更新任务生成的文件完整路径
func (m *TaskManager) UpdateTaskFilePath(taskID string, filePath string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		log.Warn("Task not found: %s", taskID)
		return
	}

	// 更新文件路径
	task.FilePath = filePath
	task.UpdatedAt = time.Now()

	log.Info("Updated task file path: %s, filePath: %s", taskID, filePath)
}

// UpdateTaskContent 更新任务缓存的文件内容
func (m *TaskManager) UpdateTaskContent(taskID string, content []byte) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		log.Warn("Task not found: %s", taskID)
		return
	}

	// 更新文件内容
	task.Content = content
	task.UpdatedAt = time.Now()

	log.Info("Updated task content: %s, size: %d bytes", taskID, len(content))
}

// UpdateTaskError 更新任务错误信息
func (m *TaskManager) UpdateTaskError(taskID string, errorMessage string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		log.Warn("Task not found: %s", taskID)
		return
	}

	// 更新错误信息
	task.ErrorMessage = errorMessage
	task.Status = TaskStatusFailed
	task.UpdatedAt = time.Now()

	log.Error("Task failed: %s, error: %s", taskID, errorMessage)
}

// GetAllTasks 获取所有任务
func (m *TaskManager) GetAllTasks() []Task {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建任务列表
	tasks := make([]Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, *task)
	}

	return tasks
}

// DeleteTask 删除任务
func (m *TaskManager) DeleteTask(taskID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 删除任务
	delete(m.tasks, taskID)

	log.Info("Deleted task: %s", taskID)
}
