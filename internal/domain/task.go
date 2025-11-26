package domain

import (
	"time"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents the priority level of a task.
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// TaskType represents the type of infrastructure task.
type TaskType string

const (
	TaskTypeConstruction TaskType = "construction"
	TaskTypeMaintenance  TaskType = "maintenance"
	TaskTypeInspection   TaskType = "inspection"
	TaskTypeEmergency    TaskType = "emergency"
)

// Task represents an infrastructure task.
type Task struct {
	ID          string       `json:"id"`
	Type        TaskType     `json:"type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Priority    TaskPriority `json:"priority"`
	Status      TaskStatus   `json:"status"`
	Location    Location     `json:"location"`
	AssignedTo  string       `json:"assigned_to"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// TaskStatusUpdate represents a task status change event.
type TaskStatusUpdate struct {
	TaskID      string     `json:"task_id"`
	CrewID      string     `json:"crew_id"`
	OldStatus   TaskStatus `json:"old_status"`
	NewStatus   TaskStatus `json:"new_status"`
	Description string     `json:"description,omitempty"`
	Timestamp   time.Time  `json:"timestamp"`
}
