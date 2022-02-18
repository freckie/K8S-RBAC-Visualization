package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/freckie/viz-rbac/internal/utils"
)

func (c *K8SClient) GetHeatmapSAResData(namespace string) (map[string]RoleRules, error) {
	cs := c.clientset
	result := make(map[string]RoleRules)

	// Get all ServiceAccounts
	saList, err := cs.CoreV1().ServiceAccounts(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, err
	}
	saNames := make([]string, len(saList.Items))
	for idx, sa := range saList.Items {
		saNames[idx] = sa.Name
	}

	// Iterate all RoleBindings
	rbList, err := cs.RbacV1().RoleBindings(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, err
	}
	for _, rb := range rbList.Items {
		if (len(rb.Subjects) == 0) || (rb.Subjects[0].Kind != "ServiceAccount") {
			continue
		}

		saName := rb.Subjects[0].Name
		if result[saName] == nil {
			result[saName] = make(RoleRules)
		}

		role, _ := c.GetRole(namespace, rb.RoleRef.Name)
		for k, v := range role {
			if result[saName][k] == nil {
				result[saName][k] = v
			} else {
				result[saName][k] = utils.ConcatString(result[saName][k], v)
			}
		}
	}

	// Iterate all ClusterRoleBindings
	crbList, err := cs.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, err
	}
	for _, crb := range crbList.Items {
		if (len(crb.Subjects) == 0) || (crb.Subjects[0].Kind != "ServiceAccount") {
			continue
		}

		saName := crb.Subjects[0].Name
		if !(utils.ContainsString(saNames, saName)) {
			continue
		}
		if result[saName] == nil {
			result[saName] = make(RoleRules)
		}

		role, _ := c.GetClusterRole(crb.RoleRef.Name)
		for k, v := range role {
			if result[saName][k] == nil {
				result[saName][k] = v
			} else {
				result[saName][k] = utils.ConcatString(result[saName][k], v)
			}
		}
	}

	return result, nil
}