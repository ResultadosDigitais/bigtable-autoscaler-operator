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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/api/option"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

// BigtableAutoscalerReconciler reconciles a BigtableAutoscaler object
type BigtableAutoscalerReconciler struct {
	ctrlclient.Client

	reader ctrlclient.Reader
	scheme *runtime.Scheme
	syncer *status.Syncer
	clock  clock.Clock
	log    logr.Logger
}

func NewBigtableReconciler(
	client ctrlclient.Client,
	reader ctrlclient.Reader,
	scheme *runtime.Scheme,
) *BigtableAutoscalerReconciler {

	log := ctrl.Log.WithName("controllers").WithName("BigtableAutoscaler")
	syncer := status.NewSyncer(client.Status(), log)

	r := &BigtableAutoscalerReconciler{
		Client: client,
		reader: reader,
		scheme: scheme,
		syncer: syncer,
		log:    log,
	}

	return r
}

func (r *BigtableAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bigtablev1.BigtableAutoscaler{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=bigtable.bigtable-autoscaler.com,resources=bigtableautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bigtable.bigtable-autoscaler.com,resources=bigtableautoscalers/status,verbs=get;update;patch
func (r *BigtableAutoscalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	r.clock = clock.RealClock{}

	var autoscaler bigtablev1.BigtableAutoscaler
	if err := r.Get(ctx, req.NamespacedName, &autoscaler); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get autoscaler: %w", err)
	}

	r.log.Info("Reconciling", "autoscaler", autoscaler.UID)

	if autoscaler.Spec.MaxScaleDownNodes == nil || *autoscaler.Spec.MaxScaleDownNodes == 0 {
		var defaultMaxScaleDownNodes int32 = 2
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

	credentialsJSON, err := r.getCredentialsJSON(ctx, autoscaler.Spec.ServiceAccountSecretRef, autoscaler.Namespace)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get credentials: %w", err)
	}

	clusterRef := autoscaler.Spec.BigtableClusterRef
	googleCloudClient, err := googlecloud.NewClientFromCredentials(ctx, credentialsJSON, clusterRef.ProjectID, clusterRef.InstanceID)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to initialize googlecloud client: %w", err)
	}

	r.syncer.Register(ctx, &autoscaler, googleCloudClient)

	desiredNodes := nodes_calculator.CalcDesiredNodes(&autoscaler.Status, &autoscaler.Spec)
	autoscaler.Status.DesiredNodes = &desiredNodes

	now := r.clock.Now()
	needUpdate := r.needUpdateNodes(&autoscaler.Status, now)
	if needUpdate {
		r.log.Info("Updating last scale time")
		autoscaler.Status.LastScaleTime = &metav1.Time{Time: now}

		r.log.Info("Metric read", "Increasing node count to", desiredNodes)
		err := scaleNodes(ctx, credentialsJSON, &clusterRef, desiredNodes)
		if err != nil {
			r.log.Error(err, "failed to update nodes")
		}
	}

	if err = r.Status().Update(ctx, &autoscaler); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update autoscaler status: %w", err)
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

func (r *BigtableAutoscalerReconciler) needUpdateNodes(status *bigtablev1.BigtableAutoscalerStatus, now time.Time) bool {
	scaleDownInterval := 1 * time.Minute

	if status.CurrentNodes == nil || status.DesiredNodes == nil {
		return false
	}

	currentNodes := *status.CurrentNodes
	desiredNodes := *status.DesiredNodes

	if desiredNodes == currentNodes {
		r.log.Info("The desired number of nodes is equal to that of the current; no need to scale nodes", "desired", desiredNodes)
		return false
	}

	if status.LastScaleTime != nil && now.Before(status.LastScaleTime.Time.Add(scaleDownInterval)) {
		r.log.Info(
			"Too soon to scale",
			"time to wait", scaleDownInterval,
			"now", now.String(),
			"last scale time", status.LastScaleTime,
			"desired", desiredNodes,
			"current", currentNodes,
		)

		return false

	}

	r.log.Info("The desired number of nodes is different than current: scaling", "desired", desiredNodes, "current", currentNodes)
	return true
}

func scaleNodes(ctx context.Context, credentialsJSON []byte, clusterRef *bigtablev1.BigtableClusterRef, desiredNodes int32) error {
	client, err := bigtable.NewInstanceAdminClient(ctx, clusterRef.ProjectID, option.WithCredentialsJSON(credentialsJSON))

	if err != nil {
		return err
	}

	return client.UpdateCluster(ctx, clusterRef.InstanceID, clusterRef.ClusterID, desiredNodes)
}
