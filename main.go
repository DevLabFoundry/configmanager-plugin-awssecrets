package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/DevLabFoundry/configmanager-plugin-awssecretsmanager/impl"
	"github.com/DevLabFoundry/configmanager/v3/config"
	"github.com/DevLabFoundry/configmanager/v3/tokenstore"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

type TokenStorePlugin struct {
	impl tokenstore.TokenStore
}

func (ts TokenStorePlugin) Value(key string, metadata []byte) (string, error) {
	return ts.impl.Value(key, metadata)
}

const PluginName string = "AWSSecretsMgr"

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

func ShowFlag(osArgs []string) bool {
	fs := flag.NewFlagSet("plugin", flag.ContinueOnError)
	vf := fs.Bool("version", false, "plugin version")
	if err := fs.Parse(osArgs); err != nil {
		return false
	}

	if *vf {
		fmt.Printf("Configmanager Plugin (%s) Version: %s-%s\n", PluginName, Version, Revision)
		return true
	}
	return false
}

func PluginSetup() (*impl.SecretsMgr, error) {

	// log set up
	log := hclog.New(hclog.DefaultOptions)

	log.SetLevel(hclog.LevelFromString("error"))

	os.Environ()
	if val, ok := os.LookupEnv(config.CONFIGMANAGER_LOG); ok && len(val) > 0 {
		if logLevel := hclog.LevelFromString(val); logLevel != hclog.NoLevel {
			log.SetLevel(logLevel)
		}
	}

	// initialize the implementation
	i, err := impl.NewSecretsMgr(context.Background(), log)
	if err != nil {
		log.Error("implementation init error", err, "impl", "awssecretsmanager")
		return nil, err
	}

	return i, nil
}

// main func should only be used for process terminantion
// Inits and serves the plugin
func main() {
	if ShowFlag(os.Args[1:]) {
		os.Exit(0)
	}

	i, err := PluginSetup()
	if err != nil {
		os.Exit(1)
	}

	// initializing the service
	ts := TokenStorePlugin{impl: i}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: tokenstore.Handshake,
		Plugins: map[string]plugin.Plugin{
			"configmanager_token_store": &tokenstore.GRPCPlugin{Impl: ts},
		},
		//
		VersionedPlugins: map[int]plugin.PluginSet{
			1: {
				"configmanager_token_store": &tokenstore.GRPCPlugin{Impl: ts},
			},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
