package planner

import (
	"encoding/base64"
	"fmt"

	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1/plan"
	"github.com/rancher/rancher/pkg/controllers/provisioningv2/rke2/machineprovision"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type s3Args struct {
	prefix      string
	secretCache corecontrollers.SecretCache
	env         bool
}

func (s *s3Args) ToArgs(s3 *rkev1.ETCDSnapshotS3, controlPlane *rkev1.RKEControlPlane) (args []string, env []string, files []plan.File, err error) {
	if s3 == nil {
		return
	}

	var (
		s3Cred s3Credential
	)

	args = append(args,
		fmt.Sprintf("--%ss3", s.prefix),
		fmt.Sprintf("--%ss3-bucket=%s", s.prefix, s3.Bucket))

	credName := s3.CloudCredentialName
	if credName == "" && controlPlane.Spec.ETCD != nil && controlPlane.Spec.ETCD.S3 != nil {
		credName = controlPlane.Spec.ETCD.S3.CloudCredentialName
	}

	s3Cred, err = getS3Credential(s.secretCache, controlPlane.Namespace, credName, s3.Region)
	if err != nil {
		return
	}

	if s3Cred.AccessKey != "" {
		args = append(args, fmt.Sprintf("--%ss3-access-key=%s", s.prefix, s3Cred.AccessKey))
	}
	if s3Cred.SecretKey != "" {
		if s.env {
			env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Cred.SecretKey))
		} else {
			args = append(args, fmt.Sprintf("--%ss3-secret-key=%s", s.prefix, s3Cred.SecretKey))
		}
	}
	if s3Cred.Region != "" {
		args = append(args, fmt.Sprintf("--%ss3-region=%s", s.prefix, s3Cred.Region))
	}
	if s3.Folder != "" {
		args = append(args, fmt.Sprintf("--%ss3-folder=%s", s.prefix, s3.Folder))
	}
	if s3.Endpoint != "" {
		args = append(args, fmt.Sprintf("--%ss3-endpoint=%s", s.prefix, s3.Endpoint))
	}
	if s3.SkipSSLVerify {
		args = append(args, fmt.Sprintf("--%ss3-skip-ssl-verify", s.prefix))
	}
	if s3.EndpointCA != "" {
		filePath := configFile(controlPlane, "s3-endpoint-ca.crt")
		files = append(files, plan.File{
			Content: base64.StdEncoding.EncodeToString([]byte(s3.EndpointCA)),
			Path:    filePath,
		})
		args = append(args, fmt.Sprintf("--%ss3-endpoint-ca=%s", s.prefix, filePath))
	}

	return
}

type s3Credential struct {
	AccessKey string
	SecretKey string
	Region    string
}

func getS3Credential(secretCache corecontrollers.SecretCache, namespace, name, region string) (result s3Credential, _ error) {
	if name == "" {
		result.Region = region
		return result, nil
	}

	secret, err := machineprovision.GetCloudCredentialSecret(secretCache, namespace, name)
	if err != nil {
		return result, fmt.Errorf("failed to lookup etcdSnapshotCloudCredentialName: %w", err)
	}

	data := map[string][]byte{}
	for k, v := range secret.Data {
		_, k = kv.RSplit(k, "-")
		data[k] = v
	}

	if region == "" {
		region = string(data["defaultRegion"])
	}
	return s3Credential{
		AccessKey: string(data["accessKey"]),
		SecretKey: string(data["secretKey"]),
		Region:    region,
	}, nil
}
