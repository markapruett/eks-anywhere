//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	cluster *framework.ClusterE2ETest
}

func (suite *Suite) SetupSuite() {
	t := suite.T()
	suite.cluster = framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
}

func getIPAddressPoolSpec(addresses []string, autoAssign bool) string {
	aList, _ := json.Marshal(addresses)
	return fmt.Sprintf(`{"addresses":%s,"autoAssign":%s,"avoidBuggyIPs":false}`, aList, strconv.FormatBool(autoAssign))
}

func getL2AdvertisementSpec(ipPoolNames []string) string {
	pools, _ := json.Marshal(ipPoolNames)
	return fmt.Sprintf(`{"ipAddressPools":%s}`, pools)
}

func getBGPAdvertisementSpec(ipPoolNames []string) string {
	pools, _ := json.Marshal(ipPoolNames)
	return fmt.Sprintf(`{"aggregationLength":32,"aggregationLengthV6":32,"ipAddressPools":%s,"localPref":123}`, pools)
}

func (suite *Suite) TestPackagesMetalLB() {
	// This should be split into multiple tests with a cluster setup in `SetupSuite`.
	// This however requires the creation of utilites managing cluster creation.
	t := suite.T()
	suite.cluster.WithCluster(func(test *framework.ClusterE2ETest) {
		kcfg := kubeconfig.FromClusterName(test.ClusterName)
		cluster := suite.cluster.Cluster()
		test.InstallCuratedPackagesController()
		ctx := context.Background()
		namespace := "metallb-system"
		test.CreateNamespace(namespace)
		packageName := "metallb"
		packageCrdName := "metallb-crds"
		packagePrefix := "test"
		test.SetPackageBundleActive()

		t.Run("Basic installation", func(t *testing.T) {
			t.Cleanup(func() {
				test.UninstallCuratedPackage(packagePrefix)
				test.UninstallCuratedPackage(packageCrdName)
			})
			test.InstallCuratedPackage(packageName, packagePrefix, kcfg)
			err := WaitForPackageToBeInstalled(test, ctx, packagePrefix, 120*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb package to be installed: %s", err)
			}
			err = test.KubectlClient.WaitForDeployment(context.Background(),
				cluster, "2m", "Available", "test-metallb-controller", namespace)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
			err = WaitForDaemonset(test, ctx, "test-metallb-speaker", namespace, 2, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
		})

		t.Run("Address pool configuration", func(t *testing.T) {
			ip := "10.100.100.1"
			ipSub := ip + "/32"
			t.Cleanup(func() {
				test.UninstallCuratedPackage(packagePrefix)
				test.UninstallCuratedPackage(packageCrdName)
			})
			test.CreateResource(ctx, fmt.Sprintf(
				`
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: test
  namespace: eksa-packages-%s
spec:
  packageName: metallb
  config: |
    IPAddressPools:
      - name: default
        addresses:
          - %s
    L2Advertisements:
      - ipAddressPools:
        - default
`, test.ClusterName, ipSub))
			err := WaitForPackageToBeInstalled(test, ctx, packagePrefix, 120*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb package to be installed: %s", err)
			}
			err = test.KubectlClient.WaitForDeployment(context.Background(),
				cluster, "2m", "Available", "test-metallb-controller", namespace)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
			err = WaitForDaemonset(test, ctx, "test-metallb-speaker", namespace, 2, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb speaker deployment to be available: %s", err)
			}

			expectedAddressPool := getIPAddressPoolSpec([]string{ipSub}, true)
			err = WaitForResource(
				test,
				ctx,
				"ipaddresspools.metallb.io/default",
				namespace,
				"{.spec}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedAddressPool),
			)

			if err != nil {
				t.Fatal(err)
			}

			expectedAdvertisement := getL2AdvertisementSpec([]string{"default"})
			err = WaitForResource(
				test,
				ctx,
				"l2advertisements.metallb.io/l2adv-0",
				namespace,
				"{.spec}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedAdvertisement),
			)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				test.KubectlClient.Delete(ctx, "service", "my-service", "default", kubeconfig.FromClusterName(test.ClusterName))
			})
			test.CreateResource(ctx, `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
`)
			err = WaitForResource(
				test,
				ctx,
				"service/my-service",
				"default",
				"{.status.loadBalancer.ingress[0].ip}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(ip),
			)
			if err != nil {
				t.Fatal(err)
			}
		})

		t.Run("BGP configuration", func(t *testing.T) {
			ip := "10.100.100.2"
			ipSub := ip + "/32"
			ipTwo := "10.100.0.1"
			ipTwoSub := ipTwo + "/32"
			t.Cleanup(func() {
				test.UninstallCuratedPackage(packagePrefix)
				test.UninstallCuratedPackage(packageCrdName)
			})
			test.CreateResource(ctx, fmt.Sprintf(
				`
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
  name: test
  namespace: eksa-packages-%s
spec:
  packageName: metallb
  config: |
    IPAddressPools:
      - name: default
        addresses:
          - %s
        autoAssign: false
      - name: bgp
        addresses:
          - %s
    L2Advertisements:
      - ipAddressPools:
        - default
    BGPAdvertisements:
      - ipAddressPools:
          - bgp
        localPref: 123
        aggregationLength: 32
        aggregationLengthV6: 32
    BGPPeers:
      - myASN: 123
        peerASN: 55001
        peerAddress: 12.2.4.2
        keepaliveTime: 30s
`, test.ClusterName, ipTwoSub, ipSub))
			err := WaitForPackageToBeInstalled(test, ctx, packagePrefix, 120*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb package to be installed: %s", err)
			}
			err = test.KubectlClient.WaitForDeployment(context.Background(),
				cluster, "2m", "Available", "test-metallb-controller", namespace)
			if err != nil {
				t.Fatalf("waiting for metallb controller deployment to be available: %s", err)
			}
			err = WaitForDaemonset(test, ctx, "test-metallb-speaker", namespace, 2, 30*time.Second)
			if err != nil {
				t.Fatalf("waiting for metallb speaker deployment to be available: %s", err)
			}

			expectedAddressPool := getIPAddressPoolSpec([]string{ipTwoSub}, false)
			err = WaitForResource(
				test,
				ctx,
				"ipaddresspools.metallb.io/default",
				namespace,
				"{.spec}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedAddressPool),
			)
			if err != nil {
				t.Fatal(err)
			}

			expectedAddressPool = getIPAddressPoolSpec([]string{ipSub}, true)
			err = WaitForResource(
				test,
				ctx,
				"ipaddresspools.metallb.io/bgp",
				namespace,
				"{.spec}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedAddressPool),
			)
			if err != nil {
				t.Fatal(err)
			}

			expectedBGPAdv := getBGPAdvertisementSpec([]string{"bgp"})
			err = WaitForResource(
				test,
				ctx,
				"bgpadvertisements.metallb.io/bgpadv-0",
				namespace,
				"{.spec}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedBGPAdv),
			)
			if err != nil {
				t.Fatal(err)
			}

			expectedBGPPeer := `{"keepaliveTime":"30s","myASN":123,"peerASN":55001,"peerAddress":"12.2.4.2","peerPort":179}`
			err = WaitForResource(
				test,
				ctx,
				"bgppeers.metallb.io/bgppeer-0",
				namespace,
				"{.spec}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(expectedBGPPeer),
			)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				test.KubectlClient.Delete(ctx, "service", "my-service", "default", kubeconfig.FromClusterName(test.ClusterName))
			})
			test.CreateResource(ctx, `
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
`)
			err = WaitForResource(
				test,
				ctx,
				"service/my-service",
				"default",
				"{.status.loadBalancer.ingress[0].ip}",
				20*time.Second,
				NoErrorPredicate,
				StringMatchPredicate(ip),
			)
			if err != nil {
				t.Fatal(err)
			}
		})
	})
}
