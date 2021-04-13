package googlecloud_test

import (
	"context"
	"errors"
	"testing"

	"bigtable-autoscaler.com/m/v2/pkg/googlecloud"

	"bigtable-autoscaler.com/m/v2/mocks"
	"github.com/stretchr/testify/mock"
)

func Test_googleCloudClient_GetCurrentCPULoad(t *testing.T) {
	mockMetricsClient := mocks.MetricClient{}
	mockTimeSeriesIterator := mocks.TimeSeriesIterator{}
	values := []int32{50, 45, 30}
	mockTimeSeriesIterator.On("Points").Return(values, nil)
	mockMetricsClient.On("ListTimeSeries", mock.Anything, mock.Anything).Return(&mockTimeSeriesIterator)

	mockMetricsClientError := mocks.MetricClient{}
	mockTimeSeriesIteratorError := mocks.TimeSeriesIterator{}
	mockTimeSeriesIteratorError.On("Points").Return(nil, errors.New("failed to get metrics"))
	mockMetricsClientError.On("ListTimeSeries", mock.Anything, mock.Anything).
		Return(&mockTimeSeriesIteratorError)

	type fields struct {
		metricsClient googlecloud.MetricClient
		projectID     string
		instanceID    string
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
				metricsClient: &mockMetricsClient,
				projectID:     "my-project-id",
				instanceID:    "my-instance-id",
				ctx:           context.Background(),
			},
			want:    50,
			wantErr: false,
		},
		{
			name: "raises error",
			fields: fields{
				metricsClient: &mockMetricsClientError,
				projectID:     "my-project-id",
				instanceID:    "my-instance-id",
				ctx:           context.Background(),
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := googlecloud.ClientBuilder(
				tt.fields.ctx,
				tt.fields.projectID,
				tt.fields.instanceID,
				tt.fields.metricsClient,
				nil,
			)
			got, err := m.GetCurrentCPULoad()
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

func Test_googleCloudClient_GetCurrentNodeCount(t *testing.T) {
	mockBigtableClient := mocks.BigtableClient{}
	mockClusterInfo := mocks.ClusterInfo{}
	clustersInfo := []googlecloud.ClusterInfo{&mockClusterInfo}
	mockClusterInfo.On("Name").Return("cluster-name-c1")
	mockClusterInfo.On("ServerNodes").Return(int32(2))
	mockBigtableClient.On("Clusters", mock.Anything, mock.Anything).Return(clustersInfo, nil)

	mockBigtableClientError := mocks.BigtableClient{}
	mockClusterInfoError := mocks.ClusterInfo{}
	clustersInfoError := []googlecloud.ClusterInfo{&mockClusterInfoError}
	mockClusterInfoError.On("Name").Return("cluster-name-c2")
	mockBigtableClientError.On("Clusters", mock.Anything, mock.Anything).Return(clustersInfoError, nil)
	type fields struct {
		bigtableClient googlecloud.BigtableClient
		projectID      string
		instanceID     string
		clusterID      string
		ctx            context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		want    int32
		wantErr bool
	}{
		{
			name: "returns the nodes count",
			fields: fields{
				bigtableClient: &mockBigtableClient,
				projectID:      "my-project-id",
				instanceID:     "my-instance-id",
				clusterID:      "cluster-name-c1",
				ctx:            context.Background(),
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "do not find cluster id",
			fields: fields{
				bigtableClient: &mockBigtableClientError,
				projectID:      "my-project-id",
				instanceID:     "my-instance-id",
				ctx:            context.Background(),
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := googlecloud.ClientBuilder(
				tt.fields.ctx,
				tt.fields.projectID,
				tt.fields.instanceID,
				nil,
				tt.fields.bigtableClient,
			)
			got, err := m.GetCurrentNodeCount(tt.fields.clusterID)
			if (err != nil) != tt.wantErr {
				t.Errorf("googleCloudClient.GetCurrentNodeCount() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("googleCloudClient.GetCurrentNodeCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
