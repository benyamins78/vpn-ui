package service

import (
	"errors"
	"time"

	"github.com/mhsanaei/3x-ui/v2/database"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/xray"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	trafficResolutionMinute  = 60
	trafficResolutionQuarter = 15 * 60
	trafficResolutionHour    = 60 * 60
	trafficHistoryRecent     = 48 * time.Hour
	trafficHistoryMedium     = 14 * 24 * time.Hour
	trafficHistoryLong       = 90 * 24 * time.Hour
)

// RecordTrafficDeltas persists the same authoritative collector deltas used by
// client_traffics. Negative values are discarded: a core restart or counter
// reset must never create a negative sample or a synthetic spike.
func RecordTrafficDeltas(tx *gorm.DB, deltas []*xray.ClientTraffic, now time.Time) error {
	if tx == nil {
		return errors.New("traffic history: nil transaction")
	}
	if now.IsZero() {
		now = time.Now()
	}
	for _, d := range deltas {
		if d == nil || d.Email == "" {
			continue
		}
		up, down := d.Up, d.Down
		if up < 0 {
			up = 0
		}
		if down < 0 {
			down = 0
		}
		if up == 0 && down == 0 {
			continue
		}
		for _, resolution := range []int{trafficResolutionMinute, trafficResolutionQuarter, trafficResolutionHour} {
			bucket := now.Unix() / int64(resolution) * int64(resolution)
			sample := &model.AccountTrafficSample{Email: d.Email, BucketAt: bucket, Resolution: resolution, Up: up, Down: down}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "email"}, {Name: "bucket_at"}, {Name: "resolution"}},
				DoUpdates: clause.Assignments(map[string]any{
					"up":   gorm.Expr("account_traffic_samples.up + ?", up),
					"down": gorm.Expr("account_traffic_samples.down + ?", down),
				}),
			}).Create(sample).Error; err != nil {
				return err
			}
		}
	}
	// Retention is enforced during normal collection, so old installations do not
	// need a separate scheduler to remain bounded.
	if err := tx.Where("resolution = ? AND bucket_at < ?", trafficResolutionMinute, now.Add(-trafficHistoryRecent).Unix()).Delete(&model.AccountTrafficSample{}).Error; err != nil {
		return err
	}
	if err := tx.Where("resolution = ? AND bucket_at < ?", trafficResolutionQuarter, now.Add(-trafficHistoryMedium).Unix()).Delete(&model.AccountTrafficSample{}).Error; err != nil {
		return err
	}
	return tx.Where("resolution = ? AND bucket_at < ?", trafficResolutionHour, now.Add(-trafficHistoryLong).Unix()).Delete(&model.AccountTrafficSample{}).Error
}

type TrafficHistoryPoint struct {
	Timestamp int64 `json:"timestamp"`
	Upload    int64 `json:"upload"`
	Download  int64 `json:"download"`
	Total     int64 `json:"total"`
}

func TrafficHistory(email string, since time.Time, now time.Time) ([]TrafficHistoryPoint, error) {
	if email == "" {
		return nil, errors.New("traffic history: email is required")
	}
	if now.IsZero() {
		now = time.Now()
	}
	if since.IsZero() || since.After(now) {
		return nil, errors.New("traffic history: invalid range")
	}
	age := now.Sub(since)
	resolution := trafficResolutionHour
	if age <= 24*time.Hour {
		resolution = trafficResolutionMinute
	} else if age <= 14*24*time.Hour {
		resolution = trafficResolutionQuarter
	}
	var rows []model.AccountTrafficSample
	err := database.GetDB().Where("email = ? AND resolution = ? AND bucket_at >= ? AND bucket_at <= ?", email, resolution, since.Unix(), now.Unix()).Order("bucket_at asc").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]TrafficHistoryPoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, TrafficHistoryPoint{Timestamp: row.BucketAt * 1000, Upload: row.Up, Download: row.Down, Total: row.Up + row.Down})
	}
	return out, nil
}
