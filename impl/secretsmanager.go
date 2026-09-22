package impl

import (
	"context"
	"encoding/json"

	"github.com/DevLabFoundry/configmanager/v3/tokenstore"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/go-hclog"
)

type secretsMgrApi interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type SecretsMgr struct {
	svc    secretsMgrApi
	ctx    context.Context
	logger hclog.Logger
}

type SecretsMgrConfig struct {
	Version string `json:"version"`
}

func NewSecretsMgr(ctx context.Context, logger hclog.Logger) (*SecretsMgr, error) {
	cfg, err := awsConf.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	c := secretsmanager.NewFromConfig(cfg)

	return &SecretsMgr{
		svc:    c,
		logger: logger,
		ctx:    ctx,
	}, nil
}

func (s *SecretsMgr) WithSvc(svc secretsMgrApi) {
	s.svc = svc
}

func (imp *SecretsMgr) Value(token string, metadata []byte) (string, error) {
	imp.logger.Info("Concrete implementation SecretsManager")
	imp.logger.Info("SecretsMgr Token: %s", token)

	storeConf := &SecretsMgrConfig{}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, storeConf); err != nil {
			imp.logger.Error("parse metadata error %v", err)
		}
	}

	version := "AWSCURRENT"
	if storeConf.Version != "" {
		version = storeConf.Version
	}

	imp.logger.Info("Getting Secret: %s @version: %s", token, version)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(token),
		VersionStage: aws.String(version),
	}

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.GetSecretValue(ctx, input)
	if err != nil {
		imp.logger.Error(tokenstore.ImplementationNetworkErr, "config.SecretMgrPrefix", err, token)
		return "", err
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	if len(result.SecretBinary) > 0 {
		return string(result.SecretBinary), nil
	}

	imp.logger.Error("value retrieved but empty for token: %v", token)
	return "", nil
}
