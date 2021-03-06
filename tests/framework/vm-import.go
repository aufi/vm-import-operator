package framework

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1 "kubevirt.io/client-go/api/v1"
	"time"

	v2vv1 "github.com/kubevirt/vm-import-operator/pkg/apis/v2v/v1beta1"
	"github.com/kubevirt/vm-import-operator/pkg/conditions"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// EnsureVMImportDoesNotExist blocks until VM import with given name does not exist in the cluster
func (f *Framework) EnsureVMImportDoesNotExist(vmiName string) error {
	return wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
		_, err := f.VMImportClient.V2vV1beta1().VirtualMachineImports(f.Namespace.Name).Get(context.TODO(), vmiName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

// EnsureVMDoesNotExist blocks until VM with given name does not exist in the cluster
func (f *Framework) EnsureVMDoesNotExist(vmName string) error {
	return wait.PollImmediate(2*time.Second, 1*time.Minute, func() (bool, error) {
		err := f.Client.Get(context.TODO(), types.NamespacedName{Namespace: f.Namespace.Name, Name: vmName}, &kubevirtv1.VirtualMachine{})
		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

// WaitForVMImportConditionInStatus blocks until VM import with given name has given status condition with given status
func (f *Framework) WaitForVMImportConditionInStatus(pollInterval time.Duration, timeout time.Duration, vmiName string, conditionType v2vv1.VirtualMachineImportConditionType, status corev1.ConditionStatus, reason string, namespace string) error {
	pollErr := wait.PollImmediate(pollInterval, timeout, func() (bool, error) {
		retrieved, err := f.VMImportClient.V2vV1beta1().VirtualMachineImports(namespace).Get(context.TODO(), vmiName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		condition := conditions.FindConditionOfType(retrieved.Status.Conditions, conditionType)
		if condition == nil {
			return false, nil
		}
		if condition.Status != status {
			return false, nil
		}
		condReason := reason
		if condReason != "" {
			if *condition.Reason != condReason {
				return false, nil
			}
		}
		return true, nil
	})
	if pollErr == wait.ErrWaitTimeout {
		retrieved, err := f.VMImportClient.V2vV1beta1().VirtualMachineImports(namespace).Get(context.TODO(), vmiName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		return fmt.Errorf(
			"Timed out waiting for the condition type 'VirtualMachineImportCondition(type: %v, reason: %v, status: %v)', got instead '%v'",
			conditionType, reason, status, retrieved.Status.Conditions,
		)
	}

	return pollErr
}

// WaitForVMToBeProcessing blocks until VM import with given name is in Processing state
func (f *Framework) WaitForVMToBeProcessing(vmiName string) error {
	return f.WaitForVMImportConditionInStatus(2*time.Second, time.Minute, vmiName, v2vv1.Processing, corev1.ConditionTrue, string(v2vv1.CopyingDisks), f.Namespace.Name)
}
