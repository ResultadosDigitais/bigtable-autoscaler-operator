/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"cloud.google.com/go/bigtable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
)

// BigtableAutoscalerReconciler reconciles a BigtableAutoscaler object
type BigtableAutoscalerReconciler struct {
	Client    ctrlclient.Client
	APIReader ctrlclient.Reader
	Log       logr.Logger
	Scheme    *runtime.Scheme

	fetcherStarted bool

	clock clock.Clock
}

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

func NewBigtableReconciler(
	Client ctrlclient.Client,
	apiReader ctrlclient.Reader,
	scheme *runtime.Scheme,
) *BigtableAutoscalerReconciler {


	r := &BigtableAutoscalerReconciler{
		Client:     Client,
		APIReader:  apiReader,
		Log:        ctrl.Log.WithName("controllers").WithName("BigtableAutoscaler"),
		Scheme:     scheme,
		fetcherStarted: false,
	}

	return r
}

// +kubebuilder:rbac:groups=bigtable.bigtable-autoscaler.com,resources=bigtableautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bigtable.bigtable-autoscaler.com,resources=bigtableautoscalers/status,verbs=get;update;patch

func (r *BigtableAutoscalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.clock = clock.RealClock{}

	autoscaler, err := r.getAutoscaler(req.NamespacedName)

	if err != nil {
		r.Log.Error(err, "failed to get autoscaler")
		return ctrl.Result{}, err
	}

	credentialsJSON, err := r.getCredentialsJSON(req.NamespacedName)

	if err != nil {
		r.Log.Error(err, "failed to get credentials")
		return ctrl.Result{}, err
	}

	if !r.fetcherStarted {
		r.fetchMetrics(credentialsJSON, req.NamespacedName)
		r.fetcherStarted = true
	}

	clusters, err := getClusters(credentialsJSON, "cdp-development", "clustering-engine")

	if err != nil {
		r.Log.Error(err, "failed to get clusters")
		return ctrl.Result{}, err
	}

	currentNodes := int32(clusters[0].ServeNodes)
	autoscaler.Status.CurrentNodes = &currentNodes
	r.Log.Info("Metric read", "node count", currentNodes)

	var defaultMaxScaleDownNodes int32 = 2

	if autoscaler.Spec.MaxScaleDownNodes == nil || *autoscaler.Spec.MaxScaleDownNodes == 0 {
		autoscaler.Spec.MaxScaleDownNodes = &defaultMaxScaleDownNodes
	}

	if autoscaler.Status.CurrentCPUUtilization == nil {
		var cpuUsage int32 = 0
		autoscaler.Status.CurrentCPUUtilization = &cpuUsage
	}

	desiredNodes := calcDesiredNodes(
		*autoscaler.Status.CurrentCPUUtilization,
		*autoscaler.Status.CurrentNodes,
		*autoscaler.Spec.TargetCPUUtilization,
		*autoscaler.Spec.MinNodes,
		*autoscaler.Spec.MaxNodes,
		*autoscaler.Spec.MaxScaleDownNodes,
	)
	autoscaler.Status.DesiredNodes = &desiredNodes

	now := r.clock.Now()
	if autoscaler.Status.LastScaleTime == nil {
		autoscaler.Status.LastScaleTime = &metav1.Time{Time: now}
	}

	needUpdate := r.needUpdateNodes(
		*autoscaler.Status.CurrentNodes,
		*autoscaler.Status.DesiredNodes,
		*autoscaler.Status.LastScaleTime,
		now)

	if needUpdate {
		r.Log.Info("Updating last scale time")
		autoscaler.Status.LastScaleTime = &metav1.Time{Time: now}
		r.Log.Info("Metric read", "Increasing node count to", desiredNodes)
		err := scaleNodes(credentialsJSON, "cdp-development", "clustering-engine", "clustering-engine-c1", desiredNodes)
		r.Log.Error(err, "failed to update nodes")
	}

	if err = r.Client.Status().Update(ctx, autoscaler); err != nil {
		if strings.Contains(err.Error(), optimisticLockErrorMsg) {
			r.Log.Info("opsi, temos um problema de concorrencia")
			return ctrl.Result{RequeueAfter: time.Second * 1}, nil
		}
		r.Log.Error(err, "failed to update autoscaler status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BigtableAutoscalerReconciler) getCredentialsJSON(namespacedName types.NamespacedName) ([]byte, error) {
	autoscaler, err := r.getAutoscaler(namespacedName)

	if err != nil {
		r.Log.Error(err, "failed to get autoscaler")
		return nil, err
	}

	ctx := context.Background()

	secretRef := autoscaler.Spec.ServiceAccountSecretRef
	var namespace string
	if secretRef.Namespace == nil || *secretRef.Namespace == "" {
		namespace = autoscaler.Namespace
	} else {
		namespace = *secretRef.Namespace
	}

	var secret corev1.Secret
	key := ctrlclient.ObjectKey{
		Name:      *secretRef.Name,
		Namespace: namespace,
	}
	if err := r.APIReader.Get(ctx, key, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(err.Error())
		}
		r.Log.Info(err.Error())
	}
	credentialsJSON := secret.Data[*secretRef.Key]

	return credentialsJSON, nil
}

func (r *BigtableAutoscalerReconciler) getAutoscaler(namespacedName types.NamespacedName) (*bigtablev1.BigtableAutoscaler, error) {
	var autoscaler bigtablev1.BigtableAutoscaler
	ctx := context.Background()

	if err := r.Client.Get(ctx, namespacedName, &autoscaler); err != nil {
		err = ctrlclient.IgnoreNotFound(err)
		if err != nil {
			r.Log.Error(err, "failed to get bigtable-autoscaler")
			return nil, err
		}
	}

	return &autoscaler, nil
}

func scaleNodes(credentialsJSON []byte, projectID, instanceID, clusterID string, desiredNodes int32) (error) {
	ctx := context.Background()

	client, err := bigtable.NewInstanceAdminClient(ctx, projectID, option.WithCredentialsJSON(credentialsJSON))

	if err != nil {
		return err
	}

	return client.UpdateCluster(ctx, instanceID, clusterID, desiredNodes)
}

func getClusters(credentialsJSON []byte, projectID, instanceID string) ([]*bigtable.ClusterInfo, error) {
	ctx := context.Background()

	client, err := bigtable.NewInstanceAdminClient(ctx, projectID, option.WithCredentialsJSON(credentialsJSON))

	if err != nil {
		return nil, err
	}

	clusters, err := client.Clusters(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func (r *BigtableAutoscalerReconciler) fetchMetrics(credentialsJSON []byte, namespacedName types.NamespacedName) {
	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)

	const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

	eg.Go(func() error {
		ticker := time.NewTicker(3 * time.Second)
		var autoscaler *bigtablev1.BigtableAutoscaler

		for {
			select {
			case <-ticker.C:
				var err error
				autoscaler, err = r.getAutoscaler(namespacedName)
				if err != nil {
					err = ctrlclient.IgnoreNotFound(err)
					if err != nil {
						r.Log.Error(err, "failed to get bigtable-autoscaler")
						return err
					}

					return nil
				}

				metric, err := getMetrics(credentialsJSON, "cdp-development")

				if err != nil {
					r.Log.Error(err, "failed to get metrics")
					return err
				}

				r.Log.V(1).Info("Metric read", "cpu utilization", metric)
				autoscaler.Status.CurrentCPUUtilization = &metric

				if err = r.Client.Status().Update(ctx, autoscaler); err != nil {
					if strings.Contains(err.Error(), optimisticLockErrorMsg) {
						r.Log.Info("opsi, temos um problema de concorrencia")
						continue
					}
					r.Log.Error(err, "failed to update autoscaler status")
					return err
				}
			}
		}
		return nil
	})
}

func getMetrics(credentialsJSON []byte, projectID string) (int32, error) {
	ctx := context.Background()

	client, err := monitoring.NewMetricClient(ctx, option.WithCredentialsJSON(credentialsJSON))

	if err != nil {
		return -1, err
	}

	startTime := time.Now().UTC().Add(time.Minute * -20)
	endTime := time.Now().UTC()
	request := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: `metric.type="bigtable.googleapis.com/cluster/cpu_load"`,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: startTime.Unix(),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: endTime.Unix(),
			},
		},
	}

	it := client.ListTimeSeries(ctx, request)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return -1, err
		}

		points := resp.Points
		return int32(rand.Intn(40) + 50), nil
		return int32(points[len(points)-1].GetValue().GetDoubleValue() * 1000000), nil
	}

	return -1, nil
}

func (r *BigtableAutoscalerReconciler) needUpdateNodes(currentNodes, desiredNodes int32, lastScaleTime metav1.Time, now time.Time) bool {
	scaleDownInterval := 1 * time.Minute
	scaleUpInterval := 1 * time.Minute

	switch {
	case desiredNodes == currentNodes:
		r.Log.V(0).Info("the desired number of nodes is equal to that of the current; no need to scale nodes")
		return false

	case desiredNodes > currentNodes && now.Before(lastScaleTime.Time.Add(scaleUpInterval)):
		r.Log.Info("too short to scale up since instance scaled nodes last",
			"now", now.String(),
			"last scale time", lastScaleTime,
		)

		return false

	case desiredNodes < currentNodes && now.Before(lastScaleTime.Time.Add(scaleDownInterval)):
		r.Log.Info("too short to scale down since instance scaled nodes last",
			"now", now.String(),
			"last scale time", lastScaleTime,
		)

		return false

	default:
		r.Log.Info("Should update nodes")
		return true
	}
}

func calcDesiredNodes(currentCPU, currentNodes, targetCPU, minNodes, maxNodes, maxScaleDownNodes int32) int32 {
	totalCPU := currentCPU * currentNodes
	desiredNodes := totalCPU/targetCPU + 1 // roundup

	if (currentNodes - desiredNodes) > maxScaleDownNodes {
		desiredNodes = currentNodes - maxScaleDownNodes
	}

	switch {
	case desiredNodes < minNodes:
		return minNodes

	case desiredNodes > maxNodes:
		return maxNodes

	default:
		return desiredNodes
	}
}

func (r *BigtableAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bigtablev1.BigtableAutoscaler{}).
		Complete(r)
}
