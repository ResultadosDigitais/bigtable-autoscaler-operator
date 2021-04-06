package googlecloud

import (
	"context"
	"testing"

	"bigtable-autoscaler.com/m/v2/mocks"
)

func Test_googleCloudClient_GetMetrics(t *testing.T) {
    metricsClient := new(mocks.MonitoringMetricClient)

	type fields struct {
		metricsClient MonitoringMetricClient
		projectID     string
		ctx           context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "a",
			fields: fields{
				metricsClient: metricsClient,
				projectID:     "my-project-id",
				ctx:           context.Background(),
			},
			want:    -1,
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
