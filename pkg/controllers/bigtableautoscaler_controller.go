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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/api/option"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/clock"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"cloud.google.com/go/bigtable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
	"bigtable-autoscaler.com/m/v2/pkg/googlecloud"
	"bigtable-autoscaler.com/m/v2/pkg/nodes_calculator"
	"bigtable-autoscaler.com/m/v2/pkg/status"
)

// BigtableAutoscalerReconciler reconciles a BigtableAutoscaler object
type BigtableAutoscalerReconciler struct {
	ctrlclient.Client

	reader  ctrlclient.Reader
	scheme  *runtime.Scheme
	log     logr.Logger
	syncers map[types.NamespacedName]*status.Syncer
	clock   clock.Clock
}

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

func NewBigtableReconciler(
	client ctrlclient.Client,
	reader ctrlclient.Reader,
	scheme *runtime.Scheme,
) *BigtableAutoscalerReconciler {

	r := &BigtableAutoscalerReconciler{
		Client:  client,
		reader:  reader,
		scheme:  scheme,
		log:     ctrl.Log.WithName("controllers").WithName("BigtableAutoscaler"),
		syncers: make(map[types.NamespacedName]*status.Syncer),
	}

	return r
}

// +kubebuilder:rbac:groups=bigtable.bigtable-autoscaler.com,resources=bigtableautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bigtable.bigtable-autoscaler.com,resources=bigtableautoscalers/status,verbs=get;update;patch

func (r *BigtableAutoscalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.clock = clock.RealClock{}

	autoscaler, err := r.getAutoscaler(ctx, req.NamespacedName)

	if err != nil {
		if errors.IsNotFound(err) {
			delete(r.syncers, req.NamespacedName)

			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}

		r.log.Error(err, "failed to get autoscaler")
		return ctrl.Result{}, err
	}

	credentialsJSON, err := r.getCredentialsJSON(ctx, autoscaler.Spec.ServiceAccountSecretRef, autoscaler.Namespace)

	if err != nil {
		r.log.Error(err, "failed to get credentials")
		return ctrl.Result{}, err
	}

	if _, ok := r.syncers[req.NamespacedName]; !ok {
		googleCloudClient, err := googlecloud.NewClientFromCredentials(ctx, credentialsJSON, "cdp-development", "clustering-engine")
		if err != nil {
			return ctrl.Result{}, err
		}

		statusSyncer := status.NewSyncer(ctx, r.Status(), autoscaler, googleCloudClient, "clustering-engine-c1", r.log)
		statusSyncer.Start()

		r.syncers[req.NamespacedName] = statusSyncer
	}

	var defaultMaxScaleDownNodes int32 = 2

	if autoscaler.Spec.MaxScaleDownNodes == nil || *autoscaler.Spec.MaxScaleDownNodes == 0 {
		autoscaler.Spec.MaxScaleDownNodes = &defaultMaxScaleDownNodes
	}

	if autoscaler.Status.CurrentCPUUtilization == nil {
		var cpuUsage int32 = 0
		autoscaler.Status.CurrentCPUUtilization = &cpuUsage
	}

	if autoscaler.Status.CurrentNodes == nil {
		var nodes int32 = 0
		autoscaler.Status.CurrentNodes = &nodes
	}

	desiredNodes := nodes_calculator.CalcDesiredNodes(&autoscaler.Status, &autoscaler.Spec)
	autoscaler.Status.DesiredNodes = &desiredNodes

	now := r.clock.Now()
	if autoscaler.Status.LastScaleTime == nil {
		autoscaler.Status.LastScaleTime = &metav1.Time{Time: now}
	}

	needUpdate := r.needUpdateNodes(
		*autoscaler.Status.CurrentNodes,
		*autoscaler.Status.DesiredNodes,
		*autoscaler.Status.LastScaleTime,
		now,
	)

	if needUpdate {
		r.log.Info("Updating last scale time")
		autoscaler.Status.LastScaleTime = &metav1.Time{Time: now}
		r.log.Info("Metric read", "Increasing node count to", desiredNodes)
		err := scaleNodes(ctx, credentialsJSON, "cdp-development", "clustering-engine", "clustering-engine-c1", desiredNodes)
		if err != nil {
			r.log.Error(err, "failed to update nodes")
		}
	}

	if err = r.Status().Update(ctx, autoscaler); err != nil {
		if strings.Contains(err.Error(), optimisticLockErrorMsg) {
			r.log.Info("opsi, temos um problema de concorrencia")
			return ctrl.Result{RequeueAfter: time.Second * 1}, nil
		}
		r.log.Error(err, "failed to update autoscaler status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BigtableAutoscalerReconciler) getCredentialsJSON(ctx context.Context, secretRef bigtablev1.ServiceAccountSecretRef, autoscalerNamespace string) ([]byte, error) {
	var namespace string

	if secretRef.Namespace == nil || *secretRef.Namespace == "" {
		namespace = autoscalerNamespace
	} else {
		namespace = *secretRef.Namespace
	}

	var secret corev1.Secret
	key := ctrlclient.ObjectKey{
		Name:      *secretRef.Name,
		Namespace: namespace,
	}

	if err := r.reader.Get(ctx, key, &secret); err != nil {
		if errors.IsNotFound(err) {
			r.log.Info(err.Error())
		}
		r.log.Info(err.Error())
	}

	credentialsJSON := secret.Data[*secretRef.Key]

	return credentialsJSON, nil
}

func (r *BigtableAutoscalerReconciler) getAutoscaler(ctx context.Context, namespacedName types.NamespacedName) (*bigtablev1.BigtableAutoscaler, error) {
	var autoscaler bigtablev1.BigtableAutoscaler

	if err := r.Get(ctx, namespacedName, &autoscaler); err != nil {
		if err != nil {
			r.log.Error(err, "failed to get bigtable-autoscaler")
			return nil, err
		}
	}

	return &autoscaler, nil
}

func scaleNodes(ctx context.Context, credentialsJSON []byte, projectID, instanceID, clusterID string, desiredNodes int32) error {
	client, err := bigtable.NewInstanceAdminClient(ctx, projectID, option.WithCredentialsJSON(credentialsJSON))

	if err != nil {
		return err
	}

	return client.UpdateCluster(ctx, instanceID, clusterID, desiredNodes)
}

func (r *BigtableAutoscalerReconciler) needUpdateNodes(currentNodes, desiredNodes int32, lastScaleTime metav1.Time, now time.Time) bool {
	scaleDownInterval := 1 * time.Minute
	scaleUpInterval := 1 * time.Minute

	switch {
	case desiredNodes == currentNodes:
		r.log.Info("the desired number of nodes is equal to that of the current; no need to scale nodes")
		return false

	case desiredNodes > currentNodes && now.Before(lastScaleTime.Time.Add(scaleUpInterval)):
		r.log.Info("too short to scale up since instance scaled nodes last",
			"now", now.String(),
			"last scale time", lastScaleTime,
		)

		return false

	case desiredNodes < currentNodes && now.Before(lastScaleTime.Time.Add(scaleDownInterval)):
		r.log.Info("too short to scale down since instance scaled nodes last",
			"now", now.String(),
			"last scale time", lastScaleTime,
		)

		return false

	default:
		r.log.Info("Should update nodes")
		return true
	}
}

func (r *BigtableAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bigtablev1.BigtableAutoscaler{}).
		Complete(r)
}
