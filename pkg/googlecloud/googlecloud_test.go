package googlecloud_test

import (
	"context"
	"errors"
	"testing"

	googlecloud "bigtable-autoscaler.com/m/v2/pkg/googlecloud"

	"bigtable-autoscaler.com/m/v2/mocks"
	"github.com/stretchr/testify/mock"
)

func Test_googleCloudClient_GetCurrentCPULoad(t *testing.T) {
	mockMetricsClientWrapper := mocks.MetricClientWrapper{}
	mockTimeSeriesIteratorWrapper := mocks.TimeSeriesIteratorWrapper{}
	values := []int32{50, 45, 30}
	mockTimeSeriesIteratorWrapper.On("Points").Return(values, nil)
	mockMetricsClientWrapper.On("ListTimeSeries", mock.Anything, mock.Anything).Return(&mockTimeSeriesIteratorWrapper)

	mockMetricsClientWrapperError := mocks.MetricClientWrapper{}
	mockTimeSeriesIteratorWrapperError := mocks.TimeSeriesIteratorWrapper{}
	mockTimeSeriesIteratorWrapperError.On("Points").Return(nil, errors.New("failed to get metrics"))
	mockMetricsClientWrapperError.On("ListTimeSeries", mock.Anything, mock.Anything).
		Return(&mockTimeSeriesIteratorWrapperError)

	type fields struct {
		metricsClient googlecloud.MetricClientWrapper
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
				metricsClient: &mockMetricsClientWrapper,
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
				metricsClient: &mockMetricsClientWrapperError,
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
	mockBigtableClientWrapper := mocks.BigtableClientWrapper{}
	mockClusterInfoWrapper := mocks.ClusterInfoWrapper{}
	clustersInfo := []googlecloud.ClusterInfoWrapper{&mockClusterInfoWrapper}
	mockClusterInfoWrapper.On("Name").Return("cluster-name-c1")
	mockClusterInfoWrapper.On("ServerNodes").Return(int32(2))
	mockBigtableClientWrapper.On("Clusters", mock.Anything, mock.Anything).Return(clustersInfo, nil)

	mockBigtableClientWrapperError := mocks.BigtableClientWrapper{}
	mockClusterInfoWrapperError := mocks.ClusterInfoWrapper{}
	clustersInfoError := []googlecloud.ClusterInfoWrapper{&mockClusterInfoWrapperError}
	mockClusterInfoWrapperError.On("Name").Return("cluster-name-c2")
	mockBigtableClientWrapperError.On("Clusters", mock.Anything, mock.Anything).Return(clustersInfoError, nil)
	type fields struct {
		bigtableClient googlecloud.BigtableClientWrapper
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
				bigtableClient: &mockBigtableClientWrapper,
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
				bigtableClient: &mockBigtableClientWrapperError,
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
