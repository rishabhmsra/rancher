//go:build (validation || infra.rke2k3s || cluster.nodedriver || extended) && !infra.any && !infra.aks && !infra.eks && !infra.gke && !infra.rke1 && !cluster.any && !cluster.custom && !sanity && !stress

package nodescaling

import (
	"testing"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/rancher/rancher/tests/framework/extensions/clusters"
	"github.com/rancher/rancher/tests/framework/extensions/provisioning"
	"github.com/rancher/rancher/tests/framework/extensions/provisioninginput"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type NodeScaleDownAndUp struct {
	suite.Suite
	session        *session.Session
	client         *rancher.Client
	ns             string
	clustersConfig *provisioninginput.Config
}

func (s *NodeScaleDownAndUp) TearDownSuite() {
	s.session.Cleanup()
}

func (s *NodeScaleDownAndUp) SetupSuite() {
	testSession := session.NewSession()
	s.session = testSession

	s.ns = provisioninginput.Namespace

	s.clustersConfig = new(provisioninginput.Config)
	config.LoadConfig(provisioninginput.ConfigurationFileKey, s.clustersConfig)

	client, err := rancher.NewClient("", testSession)
	require.NoError(s.T(), err)

	s.client = client
}

func (s *NodeScaleDownAndUp) TestEtcdScaleDownAndUp() {
	s.Run("rke2-etcd-node-scale-down-and-up", func() {
		ReplaceNodes(s.T(), s.client, s.client.RancherConfig.ClusterName, true, false, false)
	})
}

func (s *NodeScaleDownAndUp) TestControlPlaneScaleDownAndUp() {
	s.Run("rke2-controlplane-node-scale-down-and-up", func() {
		ReplaceNodes(s.T(), s.client, s.client.RancherConfig.ClusterName, false, true, false)
	})
}

func (s *NodeScaleDownAndUp) TestWorkerScaleDownAndUp() {
	s.Run("rke2-worker-node-scale-down-and-up", func() {
		ReplaceNodes(s.T(), s.client, s.client.RancherConfig.ClusterName, false, false, true)
	})
}

func (s *NodeScaleDownAndUp) TestValidate() {
	s.Run("rke2-validate", func() {
		_, stevecluster, err := clusters.GetProvisioningClusterByName(s.client, s.client.RancherConfig.ClusterName, provisioninginput.Namespace)
		require.NoError(s.T(), err)

		clusterConfig := clusters.ConvertConfigToClusterConfig(s.clustersConfig)
		provisioning.VerifyCluster(s.T(), s.client, clusterConfig, stevecluster)
	})
}

func TestScaleDownAndUp(t *testing.T) {
	suite.Run(t, new(NodeScaleDownAndUp))
}
