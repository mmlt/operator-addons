package controllers

import (
	v1alpha1 "github.com/mmlt/operator-addons/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func condition(typ v1alpha1.ClusterAddonConditionType, b bool, reason, message string) v1alpha1.ClusterAddonCondition {
	s := metav1.ConditionFalse
	if b {
		s = metav1.ConditionTrue
	}

	return v1alpha1.ClusterAddonCondition{
		Type:    typ,
		Status:  s,
		Reason:  reason,
		Message: message,
	}
}
