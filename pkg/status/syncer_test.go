package status_test

import (
	"context"
	"sync"
	"testing"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
	"bigtable-autoscaler.com/m/v2/mocks"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	ctrl "sigs.k8s.io/controller-runtime"

	"bigtable-autoscaler.com/m/v2/pkg/googlecloud"
	"bigtable-autoscaler.com/m/v2/pkg/status"
)

func TestStart(t *testing.T) {
	autoscaler := bigtablev1.BigtableAutoscaler{
		Spec: bigtablev1.BigtableAutoscalerSpec{
			BigtableClusterRef: bigtablev1.BigtableClusterRef{
				ClusterID: "cluster-id",
			},
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	mockStatusWriter := mocks.Writer{}
	mockStatusWriter.On("Update", mock.Anything, &autoscaler).Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

	cpuUsage := int32(55)
	nodesCount := int32(2)

	mockGoogleCloudClient := mocks.GoogleCloudClient{}
	mockGoogleCloudClient.On("GetCurrentCPULoad").Return(cpuUsage, nil)
	mockGoogleCloudClient.On("GetCurrentNodeCount", "cluster-id").Return(nodesCount, nil)

	type fields struct {
		ctx               context.Context
		writer            status.Writer
		autoscaler        *bigtablev1.BigtableAutoscaler
		googleCloudClient googlecloud.GoogleCloudClient
		log               logr.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "starts the syncer",
			fields: fields{
				ctx:               context.Background(),
				writer:            &mockStatusWriter,
				autoscaler:        &autoscaler,
				googleCloudClient: &mockGoogleCloudClient,
				log:               ctrl.Log.WithName("test runtime"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := status.NewSyncer(
				tt.fields.ctx,
				tt.fields.writer,
				tt.fields.log,
			)
			s.Start(
				tt.fields.autoscaler,
				tt.fields.googleCloudClient,
			)
			wg.Wait()
		})
	}
	if assert.NotNil(t, autoscaler) {
		assert.Equal(t, int32(55), *autoscaler.Status.CurrentCPUUtilization)
		assert.Equal(t, int32(2), *autoscaler.Status.CurrentNodes)
	}

}
