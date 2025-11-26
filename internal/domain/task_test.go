package domain

import (
	"testing"
	"time"
)

func TestTaskStatus(t *testing.T) {
	statuses := []TaskStatus{
		TaskStatusPending,
		TaskStatusAssigned,
		TaskStatusInProgress,
		TaskStatusCompleted,
		TaskStatusCancelled,
	}

	expected := []string{"pending", "assigned", "in_progress", "completed", "cancelled"}

	for i, status := range statuses {
		if string(status) != expected[i] {
			t.Errorf("TaskStatus[%d] = %s; want %s", i, status, expected[i])
		}
	}
}

func TestTaskPriority(t *testing.T) {
	priorities := []TaskPriority{
		TaskPriorityLow,
		TaskPriorityMedium,
		TaskPriorityHigh,
		TaskPriorityCritical,
	}

	expected := []string{"low", "medium", "high", "critical"}

	for i, priority := range priorities {
		if string(priority) != expected[i] {
			t.Errorf("TaskPriority[%d] = %s; want %s", i, priority, expected[i])
		}
	}
}

func TestTaskType(t *testing.T) {
	types := []TaskType{
		TaskTypeConstruction,
		TaskTypeMaintenance,
		TaskTypeInspection,
		TaskTypeEmergency,
	}

	expected := []string{"construction", "maintenance", "inspection", "emergency"}

	for i, taskType := range types {
		if string(taskType) != expected[i] {
			t.Errorf("TaskType[%d] = %s; want %s", i, taskType, expected[i])
		}
	}
}

func TestTask(t *testing.T) {
	now := time.Now()
	task := Task{
		ID:          "task-001",
		Type:        TaskTypeConstruction,
		Title:       "Install Power Line",
		Description: "Install new power line section A-B",
		Priority:    TaskPriorityHigh,
		Status:      TaskStatusPending,
		Location: Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Accuracy:  15.0,
		},
		AssignedTo: "crew-001",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if task.ID != "task-001" {
		t.Errorf("Task.ID = %s; want task-001", task.ID)
	}

	if task.Type != TaskTypeConstruction {
		t.Errorf("Task.Type = %s; want %s", task.Type, TaskTypeConstruction)
	}

	if task.Priority != TaskPriorityHigh {
		t.Errorf("Task.Priority = %s; want %s", task.Priority, TaskPriorityHigh)
	}
}

func TestTaskStatusUpdate(t *testing.T) {
	now := time.Now()
	update := TaskStatusUpdate{
		TaskID:      "task-001",
		CrewID:      "crew-001",
		OldStatus:   TaskStatusAssigned,
		NewStatus:   TaskStatusInProgress,
		Description: "Started work on site",
		Timestamp:   now,
	}

	if update.TaskID != "task-001" {
		t.Errorf("TaskStatusUpdate.TaskID = %s; want task-001", update.TaskID)
	}

	if update.OldStatus != TaskStatusAssigned {
		t.Errorf("TaskStatusUpdate.OldStatus = %s; want %s", update.OldStatus, TaskStatusAssigned)
	}

	if update.NewStatus != TaskStatusInProgress {
		t.Errorf("TaskStatusUpdate.NewStatus = %s; want %s", update.NewStatus, TaskStatusInProgress)
	}
}
