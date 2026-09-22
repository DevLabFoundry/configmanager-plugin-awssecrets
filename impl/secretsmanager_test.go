package impl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/DevLabFoundry/configmanager-plugin-awssecretsmanager/impl"
	"github.com/DevLabFoundry/configmanager/v3/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/go-hclog"
)

const (
	TestPhrase            string = "got: %v want: %v\n"
	TestPhraseWithContext string = "%s\n got: %v\n\n want: %v\n"
)

type mockSecretsApi func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

func (m mockSecretsApi) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

func awsSecretsMgrGetChecker(t *testing.T, params *secretsmanager.GetSecretValueInput) {
	t.Helper()
	if params.VersionStage == nil {
		t.Fatal("expect version stage to not be nil")
	}

	if strings.Contains(*params.SecretId, "#") {
		t.Errorf("incorrectly stripped token separator")
	}

	if strings.Contains(*params.SecretId, string(config.SecretMgrPrefix)) {
		t.Errorf("incorrectly stripped prefix")
	}
}

func Test_GetSecretMgr(t *testing.T) {
	var (
		tsuccessSecret = "dsgkbdsf"
	)
	tests := map[string]struct {
		token      func() string
		metadata   func() []byte
		expect     string
		mockClient func(t *testing.T) mockSecretsApi
	}{
		"successVal": {
			func() string {
				tkn, _ := config.NewParsedToken(config.SecretMgrPrefix, *config.NewConfig())
				tkn.WithSanitizedToken("/token/1")
				tkn.WithKeyPath("")
				tkn.WithMetadata("")
				return tkn.StoreToken()
			},
			func() []byte { return []byte{} },
			tsuccessSecret, func(t *testing.T) mockSecretsApi {
				return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					awsSecretsMgrGetChecker(t, params)
					return &secretsmanager.GetSecretValueOutput{
						SecretString: &tsuccessSecret,
					}, nil
				})
			},
		},
		"success with version": {
			func() string {
				tkn, _ := config.NewParsedToken(config.SecretMgrPrefix, *config.NewConfig())
				tkn.WithSanitizedToken("/token/1")
				tkn.WithKeyPath("")
				tkn.WithMetadata("")
				return tkn.StoreToken()
			},
			func() []byte {
				b, _ := json.Marshal(&impl.SecretsMgrConfig{Version: "123"})
				return b
			},
			tsuccessSecret, func(t *testing.T) mockSecretsApi {
				return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					awsSecretsMgrGetChecker(t, params)
					if *params.VersionStage != "123" {
						t.Errorf("expected version 123, got %s", *params.VersionStage)
					}
					return &secretsmanager.GetSecretValueOutput{
						SecretString: &tsuccessSecret,
					}, nil
				})
			},
		},
		"success with binary": {
			func() string {
				tkn, _ := config.NewParsedToken(config.SecretMgrPrefix, *config.NewConfig())
				tkn.WithSanitizedToken("/token/1")
				tkn.WithKeyPath("")
				tkn.WithMetadata("")
				return tkn.StoreToken()
			},
			func() []byte { return []byte{} },
			tsuccessSecret, func(t *testing.T) mockSecretsApi {
				return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					awsSecretsMgrGetChecker(t, params)
					return &secretsmanager.GetSecretValueOutput{
						SecretBinary: []byte(tsuccessSecret),
					}, nil
				})
			},
		},
		"errored": {
			func() string {
				tkn, _ := config.NewParsedToken(config.SecretMgrPrefix, *config.NewConfig())
				tkn.WithSanitizedToken("/token/1")
				tkn.WithKeyPath("")
				tkn.WithMetadata("")
				return tkn.StoreToken()
			},
			func() []byte { return []byte{} },
			"unable to retrieve secret", func(t *testing.T) mockSecretsApi {
				return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					awsSecretsMgrGetChecker(t, params)
					return nil, fmt.Errorf("unable to retrieve secret")
				})
			},
		},
		"nil to empty": {
			func() string {
				tkn, _ := config.NewParsedToken(config.SecretMgrPrefix, *config.NewConfig())
				tkn.WithSanitizedToken("/token/1")
				tkn.WithKeyPath("")
				tkn.WithMetadata("")
				return tkn.StoreToken()
			},
			func() []byte { return []byte{} },
			"", func(t *testing.T) mockSecretsApi {
				return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					awsSecretsMgrGetChecker(t, params)
					return &secretsmanager.GetSecretValueOutput{
						SecretString: nil,
					}, nil
				})
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			i, err := impl.NewSecretsMgr(context.TODO(), hclog.NewNullLogger())
			if err != nil {
				t.Errorf(TestPhrase, err.Error(), nil)
			}
			i.WithSvc(tt.mockClient(t))

			got, err := i.Value(tt.token(), tt.metadata())
			if err != nil {
				if err.Error() != tt.expect {
					t.Errorf(TestPhrase, err.Error(), tt.expect)
				}
				return
			}
			if got != tt.expect {
				t.Errorf(TestPhrase, got, tt.expect)
			}
		})
	}
}
