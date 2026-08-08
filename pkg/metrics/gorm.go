package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

const gormStartTimeKey = "metrics:gorm_start"

var (
	gormOnce        sync.Once
	dbQueriesTotal  *prometheus.CounterVec
	dbQueryDuration *prometheus.HistogramVec
)

func initGORMMetrics() {
	gormOnce.Do(func() {
		dbQueriesTotal = NewCounterVec(CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		}, "operation", "table", "status")

		dbQueryDuration = NewHistogramVec(HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, "operation", "table")
	})
}

// GORMPlugin GORM Prometheus 指标插件，采集 CRUD 执行指标。
type GORMPlugin struct{}

// NewGORMPlugin 创建 GORM 指标插件。
func NewGORMPlugin() *GORMPlugin {
	return &GORMPlugin{}
}

// Name 返回插件名称。
func (p *GORMPlugin) Name() string {
	return "prometheus_metrics"
}

// Initialize 注册 GORM CRUD 回调。
func (p *GORMPlugin) Initialize(db *gorm.DB) error {
	initGORMMetrics()
	return registerGORMCallbacks(db)
}

func registerGORMCallbacks(db *gorm.DB) error {
	ops := []struct {
		operation string
		before    func(name string, fn func(*gorm.DB)) error
		after     func(name string, fn func(*gorm.DB)) error
	}{
		{
			operation: "create",
			before: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Create().Before("gorm:create").Register(name, fn)
			},
			after: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Create().After("gorm:create").Register(name, fn)
			},
		},
		{
			operation: "read",
			before: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Query().Before("gorm:query").Register(name, fn)
			},
			after: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Query().After("gorm:query").Register(name, fn)
			},
		},
		{
			operation: "update",
			before: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Update().Before("gorm:update").Register(name, fn)
			},
			after: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Update().After("gorm:update").Register(name, fn)
			},
		},
		{
			operation: "delete",
			before: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Delete().Before("gorm:delete").Register(name, fn)
			},
			after: func(name string, fn func(*gorm.DB)) error {
				return db.Callback().Delete().After("gorm:delete").Register(name, fn)
			},
		},
	}

	for _, op := range ops {
		op := op
		if err := op.before("metrics:before_"+op.operation, gormBeforeCallback); err != nil {
			return err
		}
		if err := op.after("metrics:after_"+op.operation, func(db *gorm.DB) {
			observeGORM(db, op.operation)
		}); err != nil {
			return err
		}
	}
	return nil
}

func gormBeforeCallback(db *gorm.DB) {
	db.InstanceSet(gormStartTimeKey, time.Now())
}

func observeGORM(db *gorm.DB, operation string) {
	v, ok := db.InstanceGet(gormStartTimeKey)
	if !ok {
		return
	}
	start, ok := v.(time.Time)
	if !ok {
		return
	}

	table := db.Statement.Table
	if table == "" {
		table = "_unknown"
	}

	status := "success"
	if db.Error != nil {
		status = "error"
	}

	dbQueriesTotal.WithLabelValues(operation, table, status).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(time.Since(start).Seconds())
}
