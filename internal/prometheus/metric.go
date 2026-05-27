package prometheus

import "time"

type MetricSeries struct {
	Labels map[string]string
	Points []MetricPoint
}

type MetricPoint struct {
	Timestamp time.Time
	Value     float64
}
