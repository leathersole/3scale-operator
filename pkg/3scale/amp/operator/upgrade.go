package operator

import (
	"context"
	"fmt"
	"reflect"

	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type UpgradeApiManager struct {
	Cr              *appsv1alpha1.APIManager
	Client          client.Client
	Logger          logr.Logger
	ApiClientReader client.Reader
	Scheme          *runtime.Scheme
}

func (u *UpgradeApiManager) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeSystemAppPreHookPodEnv()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeAMPImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	if !u.highAvailabilityModeEnabled() {
		res, err = u.upgradeBackendRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemRedisImageStream()
		if res.Requeue || err != nil {
			return res, err
		}

		res, err = u.upgradeSystemDatabaseImageStream()
		if res.Requeue || err != nil {
			return res, err
		}
	}

	// TODO should there be some kind of coordination/wait between the ImageStream
	// tags creation and the trigger update/swap ??
	res, err = u.upgradeDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	// TODO because we are deprecating the use of the "latest" ImageStreamTag
	// that tag is not necessary anymore. At this moment the upgrade does NOT take
	// care of removing that ImageStreamTag although it is not used anymore.
	// Should we take care of doing that?. Keep in mind that this would probably
	// require some kind of coordination / wait to make sure we do not remove the
	// tag until the DeploymentConfigs are successfully using the new tag

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeDeploymentConfigs() (reconcile.Result, error) {
	res, err := u.upgradeAPIcastDeploymentConfigs()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *UpgradeApiManager) upgradeAPIcastDeploymentConfigs() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewApicastReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	// TODO we are only interested on the DCs reconciliation, specifically on the ImageChange trigger
	// so maybe is not a good idea to implement upgrade procedure by reconcile procedure??
	// the Reconcile method will try to reconcile everything that is reconciled in the Apicast reconciler.
	// This is, DeploymentConfigs(and not only the ImageChange trigger), ConfigMaps, Secrets, ...
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeAMPImageStreams() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewAMPImagesReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) highAvailabilityModeEnabled() bool {
	return u.Cr.Spec.HighAvailability != nil && u.Cr.Spec.HighAvailability.Enabled
}

func (u *UpgradeApiManager) upgradeBackendRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.Cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewImageStreamBaseReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr), NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.BackendImageStream())
}

func (u *UpgradeApiManager) upgradeSystemRedisImageStream() (reconcile.Result, error) {
	redis, err := Redis(u.Cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewImageStreamBaseReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr), NewImageStreamGenericReconciler())
	return reconcile.Result{}, reconciler.Reconcile(redis.SystemImageStream())
}

func (u *UpgradeApiManager) upgradeSystemDatabaseImageStream() (reconcile.Result, error) {
	if u.Cr.Spec.System.DatabaseSpec.MySQL != nil {
		return u.upgradeSystemMySQLImageStream()
	}

	if u.Cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLImageStream()
	}

	return reconcile.Result{}, fmt.Errorf("System database is not set")
}

func (u *UpgradeApiManager) upgradeSystemMySQLImageStream() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewSystemMySQLImageReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemPostgreSQLImageStream() (reconcile.Result, error) {
	baseReconciler := NewBaseReconciler(u.Client, u.ApiClientReader, u.Scheme, u.Logger)
	baseLogicReconciler := NewBaseLogicReconciler(baseReconciler)
	reconciler := NewSystemPostgreSQLImageReconciler(NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.Cr))
	return reconciler.Reconcile()
}

func (u *UpgradeApiManager) upgradeSystemAppPreHookPodEnv() (reconcile.Result, error) {
	existingDeploymentConfig := &appsv1.DeploymentConfig{}
	err := u.Client.Get(context.TODO(), types.NamespacedName{Name: "system-app", Namespace: u.Cr.Namespace}, existingDeploymentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	system, err := System(u.Cr, u.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredEnv := system.AppDeploymentConfig().Spec.Strategy.RollingParams.Pre.ExecNewPod.Env

	changed := false
	existingPreHookPod := existingDeploymentConfig.Spec.Strategy.RollingParams.Pre.ExecNewPod
	if !reflect.DeepEqual(existingPreHookPod.Env, desiredEnv) {
		existingPreHookPod.Env = desiredEnv
		changed = true
	}

	if changed {
		u.Logger.Info(fmt.Sprintf("Update object %s", ObjectInfo(existingDeploymentConfig)))
		err := u.Client.Update(context.TODO(), existingDeploymentConfig)
		return reconcile.Result{Requeue: true}, err
	}

	return reconcile.Result{}, nil
}
