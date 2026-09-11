package service

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v2/database"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/xray"
)

func TestTrafficHistoryIgnoresNegativeDeltasAndUsesTieredBuckets(t *testing.T) {
	if err := database.InitDB(filepath.Join(t.TempDir(), "history.db")); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0)
	if err := RecordTrafficDeltas(database.GetDB(), []*xray.ClientTraffic{
		{Email: "history@test", Up: 120, Down: 30}, {Email: "history@test", Up: -999, Down: -1},
	}, now); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := database.GetDB().Model(&model.AccountTrafficSample{}).Where("email = ?", "history@test").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("sample count = %d, want 3 resolutions", count)
	}
	points, err := TrafficHistory("history@test", now.Add(-time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 || points[0].Upload != 120 || points[0].Download != 30 {
		t.Fatalf("points = %+v", points)
	}
}

func TestTrafficHistoryRetentionDeletesExpiredBuckets(t *testing.T) {
	if err := database.InitDB(filepath.Join(t.TempDir(), "retention.db")); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0)
	old := &model.AccountTrafficSample{Email: "old@test", BucketAt: now.Add(-100 * 24 * time.Hour).Unix(), Resolution: trafficResolutionHour, Up: 1}
	if err := database.GetDB().Create(old).Error; err != nil {
		t.Fatal(err)
	}
	if err := RecordTrafficDeltas(database.GetDB(), []*xray.ClientTraffic{{Email: "new@test", Up: 1}}, now); err != nil {
		t.Fatal(err)
	}
	if err := database.GetDB().Where("email = ?", "old@test").First(&model.AccountTrafficSample{}).Error; err == nil {
		t.Fatal("expired hourly sample was retained")
	}
}
