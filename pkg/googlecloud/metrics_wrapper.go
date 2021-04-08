package googlecloud

import (
	"context"
	"fmt"

	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type metricClientWrapper struct {
	metricsClient *monitoring.MetricClient
}

type timeSeriesIteratorWrapper struct {
	iterator *monitoring.TimeSeriesIterator
}

func (w *timeSeriesIteratorWrapper) Points() ([]int32, error) {
	const percent float64 = 100
	ts, err := w.iterator.Next()

	if err != nil {
		return nil, fmt.Errorf("failed to iterate over time series: %w", err)
	}

	normalizePoints := make([]int32, 0)

	for _, point := range ts.Points {
		value := point.GetValue().GetDoubleValue() * percent
		normalizePoints = append(normalizePoints, int32(value))
	}

	return normalizePoints, nil
}

// Make sure the wrapper complies with its interface.
var _ interfaces.TimeSeriesIteratorWrapper = (*timeSeriesIteratorWrapper)(nil)

func (w *metricClientWrapper) ListTimeSeries(
	ctx context.Context, req *monitoringpb.ListTimeSeriesRequest,
) interfaces.TimeSeriesIteratorWrapper {
	it := w.metricsClient.ListTimeSeries(ctx, req)

	ts := timeSeriesIteratorWrapper{iterator: it}

	return &ts
}

// Make sure the wrapper complies with its interface.
var _ interfaces.MetricClientWrapper = (*metricClientWrapper)(nil)
