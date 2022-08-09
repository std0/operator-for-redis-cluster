package pod

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	rapi "github.com/IBM/operator-for-redis-cluster/api/v1alpha1"
)

// GetLabelsSet return labels associated to the redis-node pods
func GetLabelsSet(cluster *rapi.RedisCluster) (labels.Set, error) {
	desiredLabels := labels.Set{}
	if cluster == nil {
		return desiredLabels, fmt.Errorf("redisluster nil pointer")
	}
	if cluster.Spec.AdditionalLabels != nil {
		desiredLabels = cluster.Spec.AdditionalLabels
	}
	if cluster.Spec.PodTemplate != nil {
		for k, v := range cluster.Spec.PodTemplate.Labels {
			desiredLabels[k] = v
		}
	}
	desiredLabels[rapi.ClusterNameLabelKey] = cluster.Name // add redis cluster name to the Pod labels
	return desiredLabels, nil
}

// CreateRedisClusterLabelSelector creates label selector to select the jobs related to a redis cluster
func CreateRedisClusterLabelSelector(cluster *rapi.RedisCluster) (labels.Selector, error) {
	set, err := GetLabelsSet(cluster)
	if err != nil {
		return nil, err
	}
	return labels.SelectorFromSet(set), nil
}

// GetClusterAnnotationsSet return a labels.Set of annotations from the RedisCluster
func GetClusterAnnotationsSet(cluster *rapi.RedisCluster) (labels.Set, error) {
	desiredAnnotations := make(labels.Set)
	for k, v := range cluster.Annotations {
		desiredAnnotations[k] = v
	}
	return desiredAnnotations, nil
}

// GetAnnotationsSet return a labels.Set of annotations from the RedisCluster and the PodTemplate
func GetAnnotationsSet(cluster *rapi.RedisCluster) (labels.Set, error) {
	desiredAnnotations, err := GetClusterAnnotationsSet(cluster)
	if err != nil {
		return nil, err
	}

	if cluster.Spec.PodTemplate != nil {
		for k, v := range cluster.Spec.PodTemplate.Annotations {
			desiredAnnotations[k] = v
		}
	}
	return desiredAnnotations, nil
}
