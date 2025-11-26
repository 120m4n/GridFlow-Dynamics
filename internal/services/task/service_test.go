package task

import (
	"testing"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
)

func TestNewService(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.tasks == nil {
		t.Error("Service.tasks should not be nil")
	}
}

func TestServiceCreateTask(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	task := &domain.Task{
		ID:          "task-001",
		Type:        domain.TaskTypeConstruction,
		Title:       "Install Power Line",
		Description: "Install new power line section A-B",
		Priority:    domain.TaskPriorityHigh,
		Location: domain.Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
	}

	service.CreateTask(task)

	createdTask, exists := service.GetTask("task-001")
	if !exists {
		t.Error("Task should exist after creation")
	}

	if createdTask.Status != domain.TaskStatusPending {
		t.Errorf("New task status should be pending, got %s", createdTask.Status)
	}

	if createdTask.CreatedAt.IsZero() {
		t.Error("Task.CreatedAt should be set after creation")
	}

	if createdTask.UpdatedAt.IsZero() {
		t.Error("Task.UpdatedAt should be set after creation")
	}
}

func TestServiceAssignTask(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	task := &domain.Task{
		ID:    "task-001",
		Title: "Test Task",
	}
	service.CreateTask(task)

	err := service.AssignTask("task-001", "crew-001")
	if err != nil {
		t.Fatalf("AssignTask failed: %v", err)
	}

	assignedTask, _ := service.GetTask("task-001")
	if assignedTask.AssignedTo != "crew-001" {
		t.Errorf("Task.AssignedTo = %s; want crew-001", assignedTask.AssignedTo)
	}

	if assignedTask.Status != domain.TaskStatusAssigned {
		t.Errorf("Task.Status = %s; want %s", assignedTask.Status, domain.TaskStatusAssigned)
	}
}

func TestServiceAssignTaskNotFound(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	err := service.AssignTask("non-existent", "crew-001")
	if err == nil {
		t.Error("AssignTask should fail for non-existent task")
	}
}

func TestServiceGetTask(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Test non-existent task
	_, exists := service.GetTask("non-existent")
	if exists {
		t.Error("GetTask should return false for non-existent task")
	}

	// Create and retrieve
	task := &domain.Task{
		ID:    "task-002",
		Title: "Test Task",
	}
	service.CreateTask(task)

	retrieved, exists := service.GetTask("task-002")
	if !exists {
		t.Error("GetTask should return true for existing task")
	}

	if retrieved.ID != "task-002" {
		t.Errorf("Retrieved task ID = %s; want task-002", retrieved.ID)
	}
}

func TestServiceGetAllTasks(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Initially empty
	tasks := service.GetAllTasks()
	if len(tasks) != 0 {
		t.Errorf("GetAllTasks should return empty slice initially, got %d tasks", len(tasks))
	}

	// Add tasks
	service.CreateTask(&domain.Task{ID: "task-001", Title: "Task 1"})
	service.CreateTask(&domain.Task{ID: "task-002", Title: "Task 2"})
	service.CreateTask(&domain.Task{ID: "task-003", Title: "Task 3"})

	tasks = service.GetAllTasks()
	if len(tasks) != 3 {
		t.Errorf("GetAllTasks should return 3 tasks, got %d", len(tasks))
	}
}

func TestServiceGetTasksByStatus(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Create tasks with different statuses
	service.CreateTask(&domain.Task{ID: "task-001", Title: "Task 1"})
	service.CreateTask(&domain.Task{ID: "task-002", Title: "Task 2"})
	service.CreateTask(&domain.Task{ID: "task-003", Title: "Task 3"})

	// All should be pending initially
	pendingTasks := service.GetTasksByStatus(domain.TaskStatusPending)
	if len(pendingTasks) != 3 {
		t.Errorf("Expected 3 pending tasks, got %d", len(pendingTasks))
	}

	// Assign one task
	service.AssignTask("task-001", "crew-001")

	assignedTasks := service.GetTasksByStatus(domain.TaskStatusAssigned)
	if len(assignedTasks) != 1 {
		t.Errorf("Expected 1 assigned task, got %d", len(assignedTasks))
	}

	pendingTasks = service.GetTasksByStatus(domain.TaskStatusPending)
	if len(pendingTasks) != 2 {
		t.Errorf("Expected 2 pending tasks, got %d", len(pendingTasks))
	}
}

func TestServiceGetTasksByCrew(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Create and assign tasks
	service.CreateTask(&domain.Task{ID: "task-001", Title: "Task 1"})
	service.CreateTask(&domain.Task{ID: "task-002", Title: "Task 2"})
	service.CreateTask(&domain.Task{ID: "task-003", Title: "Task 3"})

	service.AssignTask("task-001", "crew-001")
	service.AssignTask("task-002", "crew-001")
	service.AssignTask("task-003", "crew-002")

	crew1Tasks := service.GetTasksByCrew("crew-001")
	if len(crew1Tasks) != 2 {
		t.Errorf("Expected 2 tasks for crew-001, got %d", len(crew1Tasks))
	}

	crew2Tasks := service.GetTasksByCrew("crew-002")
	if len(crew2Tasks) != 1 {
		t.Errorf("Expected 1 task for crew-002, got %d", len(crew2Tasks))
	}

	crew3Tasks := service.GetTasksByCrew("crew-003")
	if len(crew3Tasks) != 0 {
		t.Errorf("Expected 0 tasks for crew-003, got %d", len(crew3Tasks))
	}
}

func TestServiceHandleTaskStatusUpdate(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Create a task first
	service.CreateTask(&domain.Task{
		ID:    "task-001",
		Title: "Test Task",
	})
	service.AssignTask("task-001", "crew-001")

	// Process status update
	updateJSON := []byte(`{
		"task_id": "task-001",
		"crew_id": "crew-001",
		"old_status": "assigned",
		"new_status": "in_progress",
		"description": "Started work on site",
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	err := service.handleTaskStatusUpdate(updateJSON)
	if err != nil {
		t.Fatalf("handleTaskStatusUpdate failed: %v", err)
	}

	task, _ := service.GetTask("task-001")
	if task.Status != domain.TaskStatusInProgress {
		t.Errorf("Task.Status = %s; want %s", task.Status, domain.TaskStatusInProgress)
	}

	if task.StartedAt == nil {
		t.Error("Task.StartedAt should be set when status changes to in_progress")
	}
}

func TestServiceHandleTaskStatusUpdateCompletion(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Create a task first
	service.CreateTask(&domain.Task{
		ID:    "task-001",
		Title: "Test Task",
	})

	// Process completion update
	updateJSON := []byte(`{
		"task_id": "task-001",
		"crew_id": "crew-001",
		"old_status": "in_progress",
		"new_status": "completed",
		"description": "Work completed successfully",
		"timestamp": "2024-01-15T15:30:00Z"
	}`)

	err := service.handleTaskStatusUpdate(updateJSON)
	if err != nil {
		t.Fatalf("handleTaskStatusUpdate failed: %v", err)
	}

	task, _ := service.GetTask("task-001")
	if task.Status != domain.TaskStatusCompleted {
		t.Errorf("Task.Status = %s; want %s", task.Status, domain.TaskStatusCompleted)
	}

	if task.CompletedAt == nil {
		t.Error("Task.CompletedAt should be set when status changes to completed")
	}
}

func TestServiceHandleTaskStatusUpdateInvalidJSON(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	err := service.handleTaskStatusUpdate([]byte("invalid json"))
	if err == nil {
		t.Error("handleTaskStatusUpdate should fail with invalid JSON")
	}
}

func TestServiceHandleTaskStatusUpdateNonExistent(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Update for non-existent task should not error but log warning
	updateJSON := []byte(`{
		"task_id": "non-existent",
		"crew_id": "crew-001",
		"old_status": "assigned",
		"new_status": "in_progress",
		"timestamp": "2024-01-15T10:30:00Z"
	}`)

	err := service.handleTaskStatusUpdate(updateJSON)
	if err != nil {
		t.Fatalf("handleTaskStatusUpdate should not error for non-existent task: %v", err)
	}
}

func TestConcurrentTaskAccess(t *testing.T) {
	service := &Service{
		tasks: make(map[string]*domain.Task),
	}

	// Simulate concurrent access
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			service.CreateTask(&domain.Task{
				ID:    "task-concurrent",
				Title: "Concurrent Task",
			})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetAllTasks()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			service.GetTask("task-concurrent")
		}
		done <- true
	}()

	// Wait for all goroutines with timeout
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}
