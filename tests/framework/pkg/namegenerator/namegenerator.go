package namegenerator

import (
	"math/rand"
	"os"
	"time"

	"github.com/rancher/rancher/tests/framework/pkg/config"
)

const lowerLetterBytes = "abcdefghijklmnopqrstuvwxyz"
const upperLetterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numberBytes = "0123456789"
const defaultRandStringLength = 5
const jenkinsBuildNumKey = "JENKINS_BUILD_NUMBER"

const PytestConfigKey = "pytest"

type PytestConfig struct {
	Username          string `json:"username" yaml:"username"`
	UserToken         string `json:"userToken" yaml:"userToken"`
	ClusterNamePrefix string `json:"clusterNamePrefix" yaml:"clusterNamePrefix"`
	RunOnAllClusters  bool   `json:"runOnAllClusters" yaml:"runOnAllClusters"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringLower returns a random string with lower case alpha
// chars with the length depending on `n`. Used for creating a random string for resource names, such as clusters.
func RandStringLower(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = lowerLetterBytes[rand.Intn(len(lowerLetterBytes))]
	}
	return string(b)
}

// RandStringWithCharset returns a random string with specifc characters from the `charset` parameter
// with the length depending on `n`. Used for creating a random string for resource names, such as clusters.
func RandStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandStringLower returns a random string with all alpha-numeric chars
// with the length depending on `n`. Used for creating a random string for resource names, such as clusters.
func RandStringAll(length int) string {
	return RandStringWithCharset(length, lowerLetterBytes+upperLetterBytes+numberBytes)
}

func AppendRandomString(baseClusterName string) string {

	prefix := GetClusterNamePrefix()

	clusterName := prefix + "-" + baseClusterName + "-" + RandStringLower(defaultRandStringLength)
	return clusterName
}

func GetClusterNamePrefix() string {

	pytestConfig := new(PytestConfig)
	config.LoadConfig(PytestConfigKey, pytestConfig)

	prefix := "auto"
	jenkinsBuildNum := os.Getenv(jenkinsBuildNumKey)

	if pytestConfig.ClusterNamePrefix == "" {
		if jenkinsBuildNum != "" {
			return prefix + "-" + jenkinsBuildNum
		}
		return prefix
	}

	return pytestConfig.ClusterNamePrefix + prefix
}
