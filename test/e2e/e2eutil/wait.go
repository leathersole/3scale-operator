package e2eutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/3scale/3scale-operator/pkg/apis/capabilities/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/test"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	clientappsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	clientroutev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	corev1 "k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WaitForDeploymentConfig(t *testing.T, kubeclient kubernetes.Interface, osAppsV1Client clientappsv1.AppsV1Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		dcInterface := osAppsV1Client.DeploymentConfigs(namespace)
		dc, err := dcInterface.Get(name, metav1.GetOptions{IncludeUninitialized: true})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of '%s' DeploymentConfig\n", name)
				return false, nil
			}
			return false, err
		}

		isReady := false
		dcConditions := dc.Status.Conditions
		for _, dcCondition := range dcConditions {
			if dcCondition.Type == appsv1.DeploymentAvailable && dcCondition.Status == corev1.ConditionTrue {
				isReady = true
			}
		}
		if isReady {
			t.Logf("DeploymentConfig '%s' available\n", name)
			return true, nil
		}
		availableReplicas := dc.Status.AvailableReplicas
		desiredReplicas := dc.Spec.Replicas
		t.Logf("Waiting for full availability of %s DeploymentConfig (%d/%d)\n", name, availableReplicas, desiredReplicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func WaitForSecret(t *testing.T, kubeClient kubernetes.Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		_, secretErr := kubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		if secretErr != nil {
			if apierrors.IsNotFound(secretErr) {
				t.Logf("Waiting for availability of secret '%s'\n", name)
				return false, nil
			}
			return false, secretErr
		}

		t.Logf("Secret [%s] available\n", name)
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func WaitForReconciliationWith3scale(t *testing.T, c test.FrameworkClient, binding v1alpha1.Binding, retryInterval, timeout time.Duration) error {

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Logf("Waiting for LastSucessfulSync of binding '%s'\n", binding.Name)

		b := v1alpha1.Binding{}
		err = c.Get(context.TODO(), types.NamespacedName{Name: binding.Name, Namespace: binding.Namespace}, &b)
		if err != nil {
			return true, err
		}
		if b.GetLastSuccessfulSync() != nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil

}

func WaitForRouteFromHost(t *testing.T, kubeClient kubernetes.Interface, osRouteV1Client clientroutev1.RouteV1Interface, namespace, host string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		routeInteface := osRouteV1Client.Routes(namespace)
		routeFieldSelector := "spec.host=" + host
		routeList, err := routeInteface.List(
			metav1.ListOptions{
				IncludeUninitialized: true,
				FieldSelector:        routeFieldSelector},
		)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of Route with host '%s'\n", host)
				return false, nil
			}
			return false, err
		}

		routeItems := routeList.Items
		if len(routeItems) == 0 {
			t.Logf("Waiting for availability of Route with host '%s'\n", host)
			return false, nil
		}
		if len(routeItems) > 1 {
			return false, fmt.Errorf("Found unexpected routes with duplicated 'host' fields")
		}

		route := routeItems[0]
		routeStatusIngresses := route.Status.Ingress
		if routeStatusIngresses == nil || len(routeStatusIngresses) == 0 {
			t.Logf("Waiting for availability of Route with host '%s'\n", host)
			return false, nil
		}

		for _, routeStatusIngress := range routeStatusIngresses {
			routeStatusIngressConditions := routeStatusIngress.Conditions
			isReady := false
			for _, routeStatusIngressCondition := range routeStatusIngressConditions {
				if routeStatusIngressCondition.Type == routev1.RouteAdmitted && routeStatusIngressCondition.Status == corev1.ConditionTrue {
					isReady = true
					break
				}
			}
			if !isReady {
				t.Logf("Waiting for availability of Route with host '%s'\n", host)
				return false, nil
			}
		}

		t.Logf("Route '%s' with host '%s' available\n", route.Name, route.Spec.Host)
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}
