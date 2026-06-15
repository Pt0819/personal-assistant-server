package schedule

import (
	"context"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
)

type ScheduleService struct{}

// Create 创建日程
func (s *ScheduleService) Create(ctx context.Context, userID uint, req *CreateScheduleReq) (*model.Schedule, error) {
	schedule := &model.Schedule{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IsAllDay:    req.IsAllDay,
		Tags:        req.Tags,
		Source:      req.Source,
		Status:      "active",
	}
	if schedule.Source == "" {
		schedule.Source = "manual"
	}

	if err := global.GVA_DB.WithContext(ctx).Create(schedule).Error; err != nil {
		return nil, err
	}
	return schedule, nil
}

// Update 更新日程
func (s *ScheduleService) Update(ctx context.Context, userID uint, id uint, req *UpdateScheduleReq) (*model.Schedule, error) {
	var schedule model.Schedule
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&schedule).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if !req.StartTime.IsZero() {
		updates["start_time"] = req.StartTime
	}
	if !req.EndTime.IsZero() {
		updates["end_time"] = req.EndTime
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["is_all_day"] = req.IsAllDay

	if err := global.GVA_DB.WithContext(ctx).Model(&schedule).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload
	global.GVA_DB.WithContext(ctx).First(&schedule, id)
	return &schedule, nil
}

// Delete 软删除日程
func (s *ScheduleService) Delete(ctx context.Context, userID uint, id uint) error {
	result := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Schedule{})
	if result.RowsAffected == 0 {
		return nil // already deleted or not found
	}
	return result.Error
}

// GetByID 获取单个日程
func (s *ScheduleService) GetByID(ctx context.Context, userID uint, id uint) (*model.Schedule, error) {
	var schedule model.Schedule
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&schedule).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

// List 按时间范围获取日程列表
func (s *ScheduleService) List(ctx context.Context, userID uint, startDate, endDate time.Time) ([]model.Schedule, error) {
	var schedules []model.Schedule
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND status = 'active' AND start_time < ? AND end_time > ?", userID, endDate, startDate).
		Order("start_time ASC").
		Find(&schedules).Error
	return schedules, err
}

// CheckConflict 检测时间冲突
func (s *ScheduleService) CheckConflict(ctx context.Context, userID uint, startTime, endTime time.Time, excludeID uint) ([]model.Schedule, error) {
	var conflicts []model.Schedule
	query := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND status = 'active' AND start_time < ? AND end_time > ?", userID, endTime, startTime)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Order("start_time ASC").Find(&conflicts).Error
	return conflicts, err
}

// --- Request types ---

type CreateScheduleReq struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsAllDay    bool      `json:"is_all_day"`
	Tags        string    `json:"tags"`
	Source      string    `json:"source"`
}

type UpdateScheduleReq struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsAllDay    bool      `json:"is_all_day"`
	Tags        string    `json:"tags"`
	Status      string    `json:"status"`
}
