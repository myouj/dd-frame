package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type gormTestModel struct {
	ID   int64  `gorm:"primaryKey"`
	Name string
}

func TestGORMPluginRecordsCRUD(t *testing.T) {
	Init(Config{Enabled: true})

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Use(NewGORMPlugin()); err != nil {
		t.Fatalf("register plugin: %v", err)
	}
	if err := db.AutoMigrate(&gormTestModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	row := gormTestModel{Name: "alice"}
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := db.First(&gormTestModel{}, row.ID).Error; err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := db.Model(&row).Update("name", "bob").Error; err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := db.Delete(&row).Error; err != nil {
		t.Fatalf("delete: %v", err)
	}

	assertCounterValue(t, "db_queries_total", 4)
}

func assertCounterValue(t *testing.T, name string, want float64) {
	t.Helper()

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	var total float64
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.GetMetric() {
			total += m.GetCounter().GetValue()
		}
	}

	if total != want {
		t.Errorf("%s total = %v, want %v", name, total, want)
	}
}
