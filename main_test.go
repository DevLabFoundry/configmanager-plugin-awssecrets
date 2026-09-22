package main_test

import (
	"testing"

	main "github.com/DevLabFoundry/configmanager-plugin-awssecretsmanager"
)

func Test_PluginInit(t *testing.T) {
	t.Run("init plugin", func(t *testing.T) {
		_, err := main.PluginSetup()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func Test_ShowVersion(t *testing.T) {
	t.Run("show Version", func(t *testing.T) {
		// os.Args = append(os.Args[:1], "--version")
		if !main.ShowFlag([]string{"--version"}) {
			t.Error("should show flag")
		}
	})

	t.Run("do not ShowVersion", func(t *testing.T) {
		if main.ShowFlag([]string{"foo"}) {
			t.Error("should show flag")
		}
	})
}
