package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rancher/rancher/tests/framework/clients/rancher"
	management "github.com/rancher/rancher/tests/framework/clients/rancher/generated/management/v3"
	"github.com/rancher/rancher/tests/framework/extensions/clusters"
	"github.com/rancher/rancher/tests/framework/extensions/users"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/rancher/rancher/tests/framework/pkg/namegenerator"
	"github.com/rancher/rancher/tests/framework/pkg/session"
	"github.com/sirupsen/logrus"
)

const configFileName = "rancher_env.config"

func main() {

	pytestConfig := new(namegenerator.PytestConfig)
	config.LoadConfig(namegenerator.PytestConfigKey, pytestConfig)

	clusterNames := ""
	allAvailableClusterNames := ""

	userOwnerAccessErrClusters := ""

	client, err := rancher.NewClient("", session.NewSession())
	if err != nil {
		logrus.Fatalf("error generating steveclient: %s", err)
	}

	clustersList, err := client.Steve.SteveType(clusters.FleetSteveResourceType).List(nil)
	if err != nil {
		logrus.Fatalf("error getting cluster list: %s", err)
	}

	clusterNamePrefix := namegenerator.GetClusterNamePrefix()

	for _, cluster := range clustersList.Data {

		clusterName := cluster.Labels["management.cattle.io/cluster-display-name"]
		clusterID := cluster.Labels["management.cattle.io/cluster-name"]

		if allAvailableClusterNames != "" {
			allAvailableClusterNames += ","
		}
		allAvailableClusterNames += clusterName

		if !pytestConfig.RunOnAllClusters && !strings.HasPrefix(clusterName, clusterNamePrefix) {
			logrus.Infof("cluster name:%s does not contain prefix:%s, skipping network checks...", clusterName, clusterNamePrefix)
			continue
		}

		if clusterName == "local" {
			continue
		}

		if pytestConfig.Username != "" {

			logrus.Infof("granting cluster owner access on cluster:%s to user:%s", clusterName, pytestConfig.Username)
			err := grantClusterOwnerAccessToUser(client, clusterID, pytestConfig.Username)
			if err != nil {
				logrus.Errorf("error granting cluster owner access to cluster: %s", err)

				if userOwnerAccessErrClusters != "" {
					userOwnerAccessErrClusters += ","
				}
				userOwnerAccessErrClusters += clusterName
			}
		}

		if clusterNames != "" {
			clusterNames += ","
		}
		clusterNames += clusterName
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

	logrus.Infof("all available clusters on the instance: %s", allAvailableClusterNames)
	logrus.Info("created config file")
	logrus.Infof("conten\n: %v", content)

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
