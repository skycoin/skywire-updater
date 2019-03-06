package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/skycoin/skywire-updater/internal/pathutil"
	"github.com/skycoin/skywire-updater/pkg/update"
)

const (
	homeMode = "HOME"
)

var initConfigModes = []string{homeMode}

var (
	output  string
	replace bool
	mode    string
)

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "generates a configuration file",
	Run: func(_ *cobra.Command, _ []string) {
		output, err := filepath.Abs(output)
		if err != nil {
			log.WithError(err).Fatalln("invalid output provided")
		}
		var conf *update.Config
		var appDir string
		switch mode {
		case homeMode:
			conf = update.NewHomeConfig()
			appDir = filepath.Join(pathutil.HomeDir(), ".skycoin/skywire/apps/bin")
		default:
			log.Fatalln("invalid mode:", mode)
		}
		conf.Services.Services = map[string]*update.ServiceConfig{
			"skywire": {
				Repo:        "github.com/skycoin/skywire",
				MainBranch:  "mainnet",
				MainProcess: "skywire-node",
				Checker: update.CheckerConfig{
					Type:   update.ScriptCheckerType,
					Script: "check/bin-diff",
				},
				Updater: update.UpdaterConfig{
					Type:   update.ScriptUpdaterType,
					Script: "update/skywire",
					Envs:   []string{update.MakeEnv("APP_DIR", appDir)},
				},
			},
			"skywire-updater": {
				Repo:        "github.com/skycoin/skywire-updater",
				MainProcess: "skywire-updater",
				Checker: update.CheckerConfig{
					Type: update.GithubReleaseCheckerType,
				},
				Updater: update.UpdaterConfig{
					Type:   update.ScriptUpdaterType,
					Script: "update/to-release",
				},
			},
			"skycoin": {
				Repo:        "github.com/skycoin/skycoin",
				MainProcess: "skycoin",
				Checker: update.CheckerConfig{
					Type: update.GithubReleaseCheckerType,
				},
				Updater: update.UpdaterConfig{
					Type:   update.ScriptUpdaterType,
					Script: "update/to-release",
				},
			},
		}
		raw, err := yaml.Marshal(conf)
		if err != nil {
			log.WithError(err).Fatal("this is a bug; report to dev")
		}
		if _, err := os.Stat(output); !replace && err == nil {
			log.Fatalf("file %s already exists, stopping as 'replace,r' flag is not set", output)
		}
		if err := os.MkdirAll(filepath.Dir(output), 0750); err != nil {
			log.WithError(err).Fatalln("failed to create output directory")
		}
		if err := ioutil.WriteFile(output, raw, 0744); err != nil {
			log.WithError(err).Fatalln("failed to write file")
		}
		log.Infof("Wrote %d bytes to %s\n%s", len(raw), output, string(raw))
	},
}

func init() {
	initConfigCmd.Flags().StringVarP(&output, "output", "o", defaultConfigPaths[0], "path of output config file.")
	initConfigCmd.Flags().BoolVarP(&replace, "replace", "r", false, "whether to allow rewrite of a file that already exists.")
	initConfigCmd.Flags().StringVarP(&mode, "mode", "m", homeMode, fmt.Sprintf("config generation mode. Valid values: %v", initConfigModes))
}
