package view

import (
	"context"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
)

type ViewService struct{}

// DayView 日视图：获取指定日期的所有日程
func (s *ViewService) DayView(ctx context.Context, userID uint, date time.Time) ([]model.Schedule, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var schedules []model.Schedule
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND status = 'active' AND start_time < ? AND end_time > ?", userID, dayEnd, dayStart).
		Order("start_time ASC").
		Find(&schedules).Error
	return schedules, err
}

// WeekView 周视图：获取指定日期所在周的所有日程，按天分组
func (s *ViewService) WeekView(ctx context.Context, userID uint, date time.Time) (map[string][]model.Schedule, error) {
	// 计算本周起始（周一起始）
	weekday := date.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	weekStart := time.Date(date.Year(), date.Month(), date.Day()-int(weekday)+1, 0, 0, 0, 0, date.Location())
	weekEnd := weekStart.Add(7 * 24 * time.Hour)

	var schedules []model.Schedule
	err := global.GVA_DB.WithContext(ctx).
		Where("user_id = ? AND status = 'active' AND start_time < ? AND end_time > ?", userID, weekEnd, weekStart).
		Order("start_time ASC").
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}

	// 按天分组
	result := make(map[string][]model.Schedule)
	for i := 0; i < 7; i++ {
		dayKey := weekStart.Add(time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		result[dayKey] = []model.Schedule{}
	}

	for _, s := range schedules {
		dayKey := s.StartTime.Format("2006-01-02")
		result[dayKey] = append(result[dayKey], s)
	}

	return result, nil
}
