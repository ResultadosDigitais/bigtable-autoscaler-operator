package googlecloud

import (
	"context"
	"testing"

	"bigtable-autoscaler.com/m/v2/mocks"
	"github.com/stretchr/testify/mock"
	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
)

func Test_googleCloudClient_GetMetrics(t *testing.T) {

	mockMetricsClientWrapper := mocks.MetricClientWrapper{}
	mocksTimeSeriesIteratorWrapper := mocks.TimeSeriesIteratorWrapper{}

	mockMetricsClientWrapper.On("ListTimeSeries", mock.Anything, mock.Anything).Return(&mocksTimeSeriesIteratorWrapper)
	values := []int32{50, 45, 30}
	mocksTimeSeriesIteratorWrapper.On("Points").Return(values, nil)

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
		// TODO: Add more test cases.
		{
			name: "a",
			fields: fields{
				metricsClient: &mockMetricsClientWrapper,
				projectID:     "my-project-id",
				ctx:           context.Background(),
			},
			want:    50,
			wantErr: false,
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
