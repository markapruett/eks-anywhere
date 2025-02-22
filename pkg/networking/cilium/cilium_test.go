package cilium_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	v1alpha12 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ciliumtest struct {
	*WithT
	ctx              context.Context
	client           *mocks.MockKubernetesClient
	h                *mocks.MockHelm
	installTemplater *mocks.MockInstallTemplater
	cluster          *types.Cluster
	spec             *cluster.Spec
	ciliumValues     []byte
}

func newCiliumTest(t *testing.T) *ciliumtest {
	ctrl := gomock.NewController(t)
	h := mocks.NewMockHelm(ctrl)
	client := mocks.NewMockKubernetesClient(ctrl)
	installTemplater := mocks.NewMockInstallTemplater(ctrl)
	return &ciliumtest{
		WithT:            NewWithT(t),
		ctx:              context.Background(),
		client:           client,
		h:                h,
		installTemplater: installTemplater,
		cluster: &types.Cluster{
			Name:           "w-cluster",
			KubeconfigFile: "config.kubeconfig",
		},
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium = v1alpha1.CiliumBundle{
				Cilium: v1alpha1.Image{
					URI: "public.ecr.aws/isovalent/cilium:v1.9.13-eksa.2",
				},
				Operator: v1alpha1.Image{
					URI: "public.ecr.aws/isovalent/operator-generic:v1.9.13-eksa.2",
				},
				HelmChart: v1alpha1.Image{
					Name: "cilium-chart",
					URI:  "public.ecr.aws/isovalent/cilium:1.9.13-eksa.2",
				},
			}
			s.VersionsBundle.KubeDistro.Kubernetes.Tag = "v1.21.9-eks-1-21-10"
			s.Cluster.Spec.ClusterNetwork.CNIConfig = &v1alpha12.CNIConfig{Cilium: &v1alpha12.CiliumConfig{}}
		}),
		ciliumValues: []byte("manifest"),
	}
}
