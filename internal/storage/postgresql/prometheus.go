package postgresql

import (
	"github.com/turtacn/QuantaID/internal/metrics"
	"gorm.io/gorm"
)

// PrometheusPlugin is a GORM plugin for recording Prometheus metrics.
type PrometheusPlugin struct{}

// Name returns the name of the plugin.
func (p *PrometheusPlugin) Name() string {
	return "prometheusPlugin"
}

// Initialize initializes the plugin.
func (p *PrometheusPlugin) Initialize(db *gorm.DB) error {
	// After every query, increment the DBQueriesTotal metric.
	db.Callback().Query().After("gorm:query").Register("prometheus:after_query", func(db *gorm.DB) {
		metrics.DBQueriesTotal.Inc()
	})
	db.Callback().Create().After("gorm:create").Register("prometheus:after_create", func(db *gorm.DB) {
		metrics.DBQueriesTotal.Inc()
	})
	db.Callback().Update().After("gorm:update").Register("prometheus:after_update", func(db *gorm.DB) {
		metrics.DBQueriesTotal.Inc()
	})
	db.Callback().Delete().After("gorm:delete").Register("prometheus:after_delete", func(db *gorm.DB) {
		metrics.DBQueriesTotal.Inc()
	})
	db.Callback().Row().After("gorm:row").Register("prometheus:after_row", func(db *gorm.DB) {
		metrics.DBQueriesTotal.Inc()
	})
	db.Callback().Raw().After("gorm:raw").Register("prometheus:after_raw", func(db *gorm.DB) {
		metrics.DBQueriesTotal.Inc()
	})
	return nil
}
