package main

import (
	"fmt"
	"os"
	"time"

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

	clustersList, err := client.Management.Cluster.List(nil)
	if err != nil {
		logrus.Fatalf("error getting cluster list: %s", err)
	}

	for _, cluster := range clustersList.Data {

		if cluster.AppliedSpec.DisplayName == "local" || cluster.AppliedSpec.DisplayName == "" {
			continue
		}

		if pytestConfig.Username != "" {

			logrus.Infof("granting cluster owner access on cluster:%s to user:%s", cluster.AppliedSpec.DisplayName, pytestConfig.Username)
			err := grantClusterOwnerAccessToUser(client, cluster.ID, pytestConfig.Username)
			if err != nil {
				logrus.Errorf("error granting cluster owner access to cluster: %s", err)

				if userOwnerAccessErrClusters != "" {
					userOwnerAccessErrClusters += ","
				}
				userOwnerAccessErrClusters += cluster.AppliedSpec.DisplayName
			}
		}

		if clusterNames != "" {
			clusterNames += ","
		}
		clusterNames += cluster.AppliedSpec.DisplayName
	}

	content := ""
	cattleTestURL := fmt.Sprintf("https://%s", client.RancherConfig.Host)

	userToken := pytestConfig.UserToken
	if pytestConfig.UserToken == "" {
		userToken = client.RancherConfig.AdminToken
	}

	content = "env.CATTLE_TEST_URL='" + cattleTestURL + "'\n"
	content += "env.ADMIN_TOKEN='" + client.RancherConfig.AdminToken + "'\n"
	content += "env.USER_TOKEN='" + userToken + "'\n"
	content += "env.RANCHER_CLUSTER_NAMES='" + clusterNames + "'\n"

	if userOwnerAccessErrClusters != "" {
		content += "env.BUILD_STATE=unstable"
	}

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
		Name:            fmt.Sprintf("cluster-role-template-binding-%v", time.Now().Unix()),
		ClusterID:       clusterID,
		RoleTemplateID:  "cluster-owner",
		UserPrincipalID: fmt.Sprintf("%s://%s", "local", userId),
	})

	return err
}
