package e2e

import (
	"context"

	api "k8s.io/api/core/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"

	// for test lisibility
	. "github.com/onsi/ginkgo"
	// for test lisibility
	. "github.com/onsi/gomega"

	rapi "github.com/TheWeatherCompany/icm-redis-operator/api/v1alpha1"
	"github.com/TheWeatherCompany/icm-redis-operator/test/e2e/framework"
)

func deleteRedisCluster(kubeClient kclient.Client, cluster *rapi.RedisCluster) {
	Expect(kubeClient.Delete(context.Background(), cluster)).To(Succeed())
}

const (
	defaultPrimaries = int32(3)
	defaultReplicas  = int32(1)
)

var kubeClient kclient.Client
var cluster *rapi.RedisCluster

const clusterName = "cluster1"
const clusterNs = api.NamespaceDefault

var _ = BeforeSuite(func() {
	kubeClient = framework.BuildAndSetClients()
})

var _ = AfterSuite(func() {
	deleteRedisCluster(kubeClient, cluster)
})

var _ = Describe("RedisCluster CRUD operations", func() {
	It("should create a RedisCluster", func() {
		cluster = framework.NewRedisCluster(clusterName, clusterNs, framework.FrameworkContext.ImageTag, defaultPrimaries, defaultReplicas)
		Eventually(framework.CreateRedisNodeServiceAccountFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())

		Eventually(framework.CreateRedisClusterFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())

		Eventually(framework.IsPodDisruptionBudgetCreatedFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())

		Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

		Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "10s", "1s").ShouldNot(HaveOccurred())
	})
	Context("a RedisCluster is created", func() {
		It("should update the RedisCluster", func() {
			newTag := "new"
			cluster = framework.NewRedisCluster(clusterName, clusterNs, newTag, defaultPrimaries, defaultReplicas)

			Eventually(framework.UpdateRedisClusterFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())

			Eventually(framework.IsPodSpecUpdatedFunc(kubeClient, cluster, newTag), "5m", "5s").ShouldNot(HaveOccurred())

			Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

			Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())
		})
		It("should scale up the RedisCluster", func() {
			nbPrimary := int32(4)
			Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, &nbPrimary, nil), "5s", "1s").ShouldNot(HaveOccurred())

			Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

			Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "10s", "1s").ShouldNot(HaveOccurred())
		})
		Context("a RedisCluster is running", func() {
			When("the number of primaries is reduced", func() {
				It("should scale down the RedisCluster", func() {
					nbPrimary := int32(3)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, &nbPrimary, nil), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "10s", "1s").ShouldNot(HaveOccurred())
				})
			})
			When("the number of replicas is increased", func() {
				It("should create additional replicas for each primary in the RedisCluster", func() {
					replicas := int32(2)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, nil, &replicas), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "10s", "1s").ShouldNot(HaveOccurred())
				})
			})
			When("the number of replicas is decreased", func() {
				It("should delete replicas for each primary in the RedisCluster", func() {
					replicas := int32(1)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, nil, &replicas), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "10s", "1s").ShouldNot(HaveOccurred())
				})
			})
			When("the number of primaries is decreased and the number of replicas is increased", func() {
				It("should scale down the primaries and create additional replicas in the RedisCluster", func() {
					nbPrimary := int32(2)
					replicas := int32(2)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, &nbPrimary, &replicas), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())
				})
			})
			When("the number of primaries is increased and the number of replicas is decreased", func() {
				It("should scale up the primaries and delete replicas in the RedisCluster", func() {
					nbPrimary := int32(3)
					replicas := int32(1)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, &nbPrimary, &replicas), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())
				})
			})
			When("the number of primaries is increased and the number of replicas is increased", func() {
				It("should scale up the primaries and create additional replicas in the RedisCluster", func() {
					nbPrimary := int32(4)
					replicas := int32(2)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, &nbPrimary, &replicas), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())
				})
			})
			When("the number of primaries is decreased and the number of replicas is decreased", func() {
				It("should scale down the primaries and delete replicas in the RedisCluster", func() {
					nbPrimary := int32(3)
					replicas := int32(1)
					Eventually(framework.UpdateConfigRedisClusterFunc(kubeClient, cluster, &nbPrimary, &replicas), "5s", "1s").ShouldNot(HaveOccurred())

					Eventually(framework.IsRedisClusterStartedFunc(kubeClient, cluster), "5m", "5s").ShouldNot(HaveOccurred())

					Eventually(framework.ZonesBalancedFunc(kubeClient, cluster), "5s", "1s").ShouldNot(HaveOccurred())
				})
			})
		})
	})

})
