package metrics

type MetricsClient interface {
	GetMetrics() (int32, error)
}
