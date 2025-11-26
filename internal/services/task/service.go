// Package task provides the task management microservice.
package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/internal/domain"
	"github.com/120m4n/GridFlow-Dynamics/internal/messaging"
)

// Service handles task management operations.
type Service struct {
	publisher *messaging.Publisher
	consumer  *messaging.Consumer
	tasks     map[string]*domain.Task
	mu        sync.RWMutex
}

// NewService creates a new task management service.
func NewService(publisher *messaging.Publisher, consumer *messaging.Consumer) *Service {
	return &Service{
		publisher: publisher,
		consumer:  consumer,
		tasks:     make(map[string]*domain.Task),
	}
}

// Start initializes and starts the task service.
func (s *Service) Start() error {
	err := s.consumer.BindQueue(
		messaging.ExchangeTaskEvents,
		messaging.RoutingKeyTaskStatus,
		s.handleTaskStatusUpdate,
	)
	if err != nil {
		return fmt.Errorf("failed to bind task status updates: %w", err)
	}

	return s.consumer.Start()
}

func (s *Service) handleTaskStatusUpdate(body []byte) error {
	var update domain.TaskStatusUpdate
	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("failed to unmarshal task status update: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[update.TaskID]
	if !exists {
		log.Printf("Task %s not found for status update", update.TaskID)
		return nil
	}

	task.Status = update.NewStatus
	task.UpdatedAt = update.Timestamp

	if update.NewStatus == domain.TaskStatusInProgress && task.StartedAt == nil {
		task.StartedAt = &update.Timestamp
	}

	if update.NewStatus == domain.TaskStatusCompleted {
		task.CompletedAt = &update.Timestamp
	}

	log.Printf("Task %s status updated: %s -> %s by crew %s",
		update.TaskID, update.OldStatus, update.NewStatus, update.CrewID)

	return nil
}

// CreateTask creates a new task.
func (s *Service) CreateTask(task *domain.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Status = domain.TaskStatusPending
	s.tasks[task.ID] = task
}

// AssignTask assigns a task to a crew.
func (s *Service) AssignTask(taskID, crewID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	task.AssignedTo = crewID
	task.Status = domain.TaskStatusAssigned
	task.UpdatedAt = time.Now()

	return nil
}

// PublishTaskStatusUpdate publishes a task status update event.
func (s *Service) PublishTaskStatusUpdate(ctx context.Context, update domain.TaskStatusUpdate) error {
	return s.publisher.Publish(ctx, messaging.ExchangeTaskEvents, messaging.RoutingKeyTaskStatus, update)
}

// UpdateTaskStatus updates and publishes a task status change.
func (s *Service) UpdateTaskStatus(ctx context.Context, taskID, crewID string, newStatus domain.TaskStatus, description string) error {
	s.mu.RLock()
	task, exists := s.tasks[taskID]
	if !exists {
		s.mu.RUnlock()
		return fmt.Errorf("task %s not found", taskID)
	}
	oldStatus := task.Status
	s.mu.RUnlock()

	update := domain.TaskStatusUpdate{
		TaskID:      taskID,
		CrewID:      crewID,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
		Description: description,
		Timestamp:   time.Now(),
	}

	return s.PublishTaskStatusUpdate(ctx, update)
}

// GetTask returns a task by ID.
func (s *Service) GetTask(taskID string) (*domain.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[taskID]
	return task, exists
}

// GetAllTasks returns all tasks.
func (s *Service) GetAllTasks() []*domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*domain.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetTasksByStatus returns tasks filtered by status.
func (s *Service) GetTasksByStatus(status domain.TaskStatus) []*domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*domain.Task
	for _, task := range s.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetTasksByCrew returns tasks assigned to a specific crew.
func (s *Service) GetTasksByCrew(crewID string) []*domain.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var tasks []*domain.Task
	for _, task := range s.tasks {
		if task.AssignedTo == crewID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// Stop stops the task service.
func (s *Service) Stop() error {
	return s.consumer.Stop()
}
