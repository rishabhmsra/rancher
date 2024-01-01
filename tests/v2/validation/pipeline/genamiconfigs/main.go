package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rancher/rancher/tests/framework/extensions/machinepools"
	"github.com/rancher/rancher/tests/framework/extensions/rke1/nodetemplates"
	"github.com/rancher/rancher/tests/framework/pkg/config"
	"github.com/sirupsen/logrus"
)

const (
	pairSeparator      = ","
	userAndIdSeparator = ":"

	dirName              = "cattle-configs"
	configEnvironmentKey = "CATTLE_TEST_CONFIG"
	awsFilePrefix        = "aws"
	amis                 = "amis"
)

type amiInfo struct {
	Id      string
	SshUser string
}

type Amis struct {
	AWS AwsAMIsConfig `json:"aws" yaml:"aws"`
}

type AwsAMIsConfig struct {
	IdUserPairs string `json:"idUserPairs" yaml:"idUserPairs"` // value needs to be in form amiID1:sshUser1,amiID2:sshUser2 and so on
}

func main() {

	configPath := os.Getenv(configEnvironmentKey)

	//make cattle-configs dir
	err := config.NewConfigurationsDir(dirName)
	if err != nil {
		logrus.Fatal("error while creating configs dir", err)
	}

	amisConfig := new(Amis)
	config.LoadConfig(amis, amisConfig)

	awsAMIsInfo, err := getAWSAMIsInfo(&amisConfig.AWS)
	if err != nil {
		log.Fatalf(`aws amis Info is not in correct format, correct format:  (amiID1:sshUser1,amiID2:sshUser2 and so on)
		given string: %s`, amisConfig.AWS.IdUserPairs)
	}

	copiedConfig, err := os.ReadFile(configPath)
	if err != nil {
		logrus.Fatal("error while copying upgrade config", err)
	}

	if len(awsAMIsInfo) == 0 {

		newConfigName := config.NewConfigFileName(dirName, "cattle-config")

		err = newConfigName.NewFile(copiedConfig)
		if err != nil {
			logrus.Infof("error creating config flie: %v", err)
		}

		logrus.Infof("created config file: %s", newConfigName)

		return
	}

	for _, amiInfo := range awsAMIsInfo {

		newConfigName := config.NewConfigFileName(dirName, awsFilePrefix, amiInfo.Id, amiInfo.SshUser)

		newConfigName.NewFile(copiedConfig)

		logrus.Infof("created config file: %s", newConfigName)

		newConfigName.SetEnvironmentKey()

		awsNodeTemplateConfig := new(nodetemplates.AmazonEC2NodeTemplateConfig)

		config.LoadAndUpdateConfig(nodetemplates.AmazonEC2NodeTemplateConfigurationFileKey, awsNodeTemplateConfig, func() {
			awsNodeTemplateConfig.AMI = amiInfo.Id
			awsNodeTemplateConfig.SSHUser = amiInfo.SshUser
		})

		machineConfig := new(machinepools.AWSMachineConfig)
		config.LoadAndUpdateConfig(machinepools.AWSMachineConfigConfigurationFileKey, machineConfig, func() {
			machineConfig.AMI = amiInfo.Id
			machineConfig.SSHUser = amiInfo.SshUser
		})
	}
}

func getAWSAMIsInfo(awsAMIsConfig *AwsAMIsConfig) ([]amiInfo, error) {

	if awsAMIsConfig == nil || awsAMIsConfig.IdUserPairs == "" {
		return []amiInfo{}, nil
	}

	pairs := strings.Split(awsAMIsConfig.IdUserPairs, ",")
	amisInfo := []amiInfo{}

	for _, pair := range pairs {

		values := strings.Split(pair, userAndIdSeparator)
		if len(values) != 2 {
			return amisInfo, fmt.Errorf("pair: (%s) is not in correct format", pair)
		}
		amisInfo = append(amisInfo, amiInfo{Id: values[0], SshUser: values[1]})
	}

	return amisInfo, nil
}
