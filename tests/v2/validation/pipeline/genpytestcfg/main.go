package main

import (
	"fmt"
	"os"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	management "github.com/rancher/rancher/tests/framework/clients/rancher/generated/management/v3"
	"github.com/rancher/rancher/tests/framework/extensions/users"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/sirupsen/logrus"
)

const pytestConfigKey = "pytest"
const configFileName = "rancher_env.config"

type PytestConfig struct {
	Username  string `json:"username" yaml:"username"`
	UserToken string `json:"userToken" yaml:"userToken"`
}

func main() {

	pytestConfig := new(PytestConfig)
	config.LoadConfig(pytestConfigKey, pytestConfig)

	clusterNames := ""
	userOwnerAccessErrClusters := ""

	client, err := rancher.NewClient("", session.NewSession())
	if err != nil {
		logrus.Fatalf("error generating steveclient: %s", err)
	}

	clustersList, err := client.Steve.SteveType("management.cattle.io.cluster").List(nil)
	if err != nil {
		logrus.Fatalf("error getting cluster list: %s", err)
	}

	for _, cluster := range clustersList.Data {

		if cluster.Name == "local" {
			continue
		}

		if pytestConfig.Username != "" {

			logrus.Infof("granting cluster owner access on cluster:%s to user:%s", cluster.Name, pytestConfig.Username)
			err := grantClusterOwnerAccessToUser(client, cluster.ID, pytestConfig.Username)
			if err != nil {
				logrus.Errorf("error granting cluster owner access to cluster: %s", err)

				if userOwnerAccessErrClusters != "" {
					userOwnerAccessErrClusters += ","
				}
				userOwnerAccessErrClusters += cluster.Name
			}
		}

		if clusterNames != "" {
			clusterNames += ","
		}
		clusterNames += cluster.Name
	}

	content := ""
	cattleTestURL := fmt.Sprintf("https://%s", client.RancherConfig.Host)

	userToken := pytestConfig.UserToken
	if pytestConfig.UserToken != "" {
		userToken = client.RancherConfig.AdminToken
	}

	content = "env.CATTLE_TEST_URL='" + cattleTestURL + "'\n"
	content += "env.ADMIN_TOKEN='" + client.RancherConfig.AdminToken + "'\n"
	content += "env.USER_TOKEN='" + userToken + "'\n"
	content += "env.RANCHER_CLUSTER_NAMES='" + clusterNames + "'\n"
	content += "env.PYTEST_OPTIONS='" + `-k "test_wl or test_connectivity or test_ingress or test_service_discovery or test_websocket"` + "'\n"

	if userOwnerAccessErrClusters != "" {
		content += "env.BUILD_STATE=unstable"
	}

	wrkDir, err := os.Getwd()
	if err != nil {
		logrus.Errorf("error geting current working directory")
	}

	logrus.Infof("current working directory: %s", wrkDir)

	file, err := os.Create(configFileName)
	if err != nil {
		logrus.Errorf("failed to create config file: %s", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		logrus.Errorf("falied to write to config file:  %s", err)
	}

	logrus.Infof("created config file")

	if userOwnerAccessErrClusters != "" {
		logrus.Info("****************************************************")
		logrus.Infof("error granting cluster access owner to user: %s on clusters: %s", pytestConfig.Username, userOwnerAccessErrClusters)
		logrus.Info("****************************************************")
	}
}

func grantClusterOwnerAccessToUser(client *rancher.Client, clusterID string, username string) error {

	userId, err := users.GetUserIDByName(client, username)
	if err != nil {
		return err
	}

	_, err = client.Management.ClusterRoleTemplateBinding.Create(&management.ClusterRoleTemplateBinding{
		Name:            "cluster-role-template-binding-1",
		ClusterID:       clusterID,
		RoleTemplateID:  "cluster-owner",
		UserPrincipalID: fmt.Sprintf("%s://%s", "local", userId),
	})

	return err
}
