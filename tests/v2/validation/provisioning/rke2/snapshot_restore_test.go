package rke2

import (
	"context"
	"fmt"

	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/rancher/tests/framework/clients/rancher"
	"github.com/rancher/rancher/tests/framework/extensions/cloudcredentials"
	"github.com/rancher/rancher/tests/framework/extensions/clusters"
	"github.com/rancher/rancher/tests/framework/extensions/machinepools"
	"github.com/rancher/rancher/tests/framework/extensions/secrets"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/rancher/tests/framework/pkg/namegenerator"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/rancher/rancher/tests/framework/pkg/wait"
	"github.com/rancher/rancher/tests/integration/pkg/defaults"
	provisioning "github.com/rancher/rancher/tests/v2/validation/provisioning"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RKE2EtcdSnapshotRestoreTestSuite struct {
	suite.Suite
	session     *session.Session
	client      *rancher.Client
	config      *rancher.Config
	clusterName string
	namespace   string
}

var phases = []rkev1.ETCDSnapshotPhase{
	rkev1.ETCDSnapshotPhaseStarted,
	rkev1.ETCDSnapshotPhaseShutdown,
	rkev1.ETCDSnapshotPhaseRestore,
	rkev1.ETCDSnapshotPhaseRestartCluster,
	rkev1.ETCDSnapshotPhaseFinished,
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) TearDownSuite() {
	r.session.Cleanup()
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) SetupSuite() {
	testSession := session.NewSession(r.T())
	r.session = testSession

	r.config = new(rancher.Config)
	config.LoadConfig(provisioning.ConfigurationFileKey, r.config)

	client, err := rancher.NewClient("", testSession)
	require.NoError(r.T(), err)

	r.client = client

	r.clusterName = r.client.RancherConfig.ClusterName
	r.namespace = r.client.RancherConfig.ClusterName
}

func (r *RKE2EtcdSnapshotRestoreTestSuite) TestEtcdSnapshotRestoreFreshCluster(provider Provider, kubeVersion string, nodesAndRoles []machinepools.NodeRoles, credential *cloudcredentials.CloudCredential) {
	name := fmt.Sprintf("Provider_%s/Kubernetes_Version_%s/Nodes_%v", provider.Name, kubeVersion, nodesAndRoles)

	r.Run(name, func() {
		testSession := session.NewSession(r.T())
		defer testSession.Cleanup()

		testSessionClient, err := r.client.WithSession(testSession)
		require.NoError(r.T(), err)

		clusterName := provisioning.AppendRandomString(fmt.Sprintf("%s-%s", r.clusterName, provider.Name))
		generatedPoolName := fmt.Sprintf("nc-%s-pool1-", clusterName)
		machinePoolConfig := provider.MachinePoolFunc(generatedPoolName, namespace)

		machineConfigResp, err := machinepools.CreateMachineConfig(provider.MachineConfig, machinePoolConfig, testSessionClient)

		machinePools := machinepools.RKEMachinePoolSetup(nodesAndRoles, machineConfigResp)

		cluster := clusters.NewRKE2ClusterConfig(clusterName, namespace, "calico", credential.ID, kubeVersion, machinePools)

		//if strings.Contains(kubeVersion, "k3s") {
		//	cluster.Spec.RKEConfig.MachineGlobalConfig.SecretsEncryption = true
		//}

		clusterResp, err := clusters.CreateRKE2Cluster(testSessionClient, cluster)
		require.NoError(r.T(), err)

		kubeRKEClient, err := r.client.GetKubeAPIProvisioningClient()
		require.NoError(r.T(), err)

		result, err := kubeRKEClient.RKEControlPlanes(namespace).Watch(context.TODO(), metav1.ListOptions{
			FieldSelector:  "metadata.name=" + cluster.ID,
			TimeoutSeconds: &defaults.WatchTimeoutSeconds,
		})
		require.NoError(r.T(), err)

		checkFunc := clusters.IsProvisioningClusterReady

		err = wait.WatchWait(result, checkFunc)
		assert.NoError(r.T(), err)
		assert.Equal(r.T(), clusterName, clusterResp.ObjectMeta.Name)

		//require.NoError(r.T(), r.rotateEncryptionKeys(cluster.ID, 1))
		//logrus.Infof("Successfully completed encryption key rotation for %s", name)

		// create 1000 secrets
		for i := 0; i < 1000; i++ {
			s := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: fmt.Sprintf("etcd-snapshot-rotation-test-%d-", i),
				},
				Data: map[string][]byte{
					"key": []byte(namegenerator.RandStringLower(5)),
				},
			}

			_, err = secrets.CreateSecret(r.client, s, cluster.ID, "default")
			require.NoError(r.T(), err)
		}

		//require.NoError(r.T(), r.rotateEncryptionKeys(cluster.ID, 2))
		//logrus.Infof("Successfully completed second encryption key rotation for %s", name)
	})
}
