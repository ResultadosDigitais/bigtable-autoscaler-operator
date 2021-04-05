package googlecloud

type GoogleCloudClient interface {
	GetMetrics() (int32, error)
}
