package googlecloud

import (
	"context"
	"errors"
	"testing"

	"bigtable-autoscaler.com/m/v2/mocks"
	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"github.com/stretchr/testify/mock"
)

func Test_googleCloudClient_GetMetrics(t *testing.T) {

	mockMetricsClientWrapper := mocks.MetricClientWrapper{}
	mocksTimeSeriesIteratorWrapper := mocks.TimeSeriesIteratorWrapper{}
	values := []int32{50, 45, 30}
	mocksTimeSeriesIteratorWrapper.On("Points").Return(values, nil)
	mockMetricsClientWrapper.On("ListTimeSeries", mock.Anything, mock.Anything).Return(&mocksTimeSeriesIteratorWrapper)

	mockMetricsClientWrapperError := mocks.MetricClientWrapper{}
	mocksTimeSeriesIteratorWrapperError := mocks.TimeSeriesIteratorWrapper{}
	mocksTimeSeriesIteratorWrapperError.On("Points").Return(nil, errors.New("Failed to get metrics"))
	mockMetricsClientWrapperError.On("ListTimeSeries", mock.Anything, mock.Anything).Return(&mocksTimeSeriesIteratorWrapperError)

	type fields struct {
		metricsClient interfaces.MetricClientWrapper
		projectID     string
		ctx           context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		want    int32
		wantErr bool
	}{
		{
			name: "returns the first value of the series",
			fields: fields{
				metricsClient: &mockMetricsClientWrapper,
				projectID:     "my-project-id",
				ctx:           context.Background(),
			},
			want:    50,
			wantErr: false,
		},
		{
			name: "raises error",
			fields: fields{
				metricsClient: &mockMetricsClientWrapperError,
				projectID:     "my-project-id",
				ctx:           context.Background(),
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &googleCloudClient{
				metricsClient: tt.fields.metricsClient,
				projectID:     tt.fields.projectID,
				ctx:           tt.fields.ctx,
			}
			got, err := m.GetMetrics()
			if (err != nil) != tt.wantErr {
				t.Errorf("googleCloudClient.GetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("googleCloudClient.GetMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}
