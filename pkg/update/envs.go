package update

import (
	"fmt"
	"os"
)

const (
	// EnvRepo can be used by checkers or updaters to determine the repository
	// URL of the service.
	EnvRepo = "SWU_REPO"

	// EnvMainBranch can be used by checkers or updaters to determine the main
	// branch fo the service.
	EnvMainBranch = "SWU_MAIN_BRANCH"

	// EnvMainProcess can be used by checkers or updaters to determine the main
	// process name for the service.
	EnvMainProcess = "SWU_MAIN_PROCESS"

	// EnvBinDir can be used by checkers or updaters to determine the path in
	// which to install the binaries.
	EnvBinDir = "SWU_BIN_DIR"

	// EnvToVersion can be used by updaters to determine the version to update
	// the service to.
	EnvToVersion = "SWU_TO_VERSION"

	// EnvGithubUsername can be used by checkers or updaters for github
	// authentication (needs to be set manually).
	EnvGithubUsername = "SWU_GITHUB_USERNAME"

	// EnvGithubAccessToken can be used by checkers or updaters scripts for
	// github authentication (needs to be set manually).
	EnvGithubAccessToken = "SWU_GITHUB_ACCESS_TOKEN" //nolint:gosec
)

// MakeEnv makes an environment variable string of format '<key>=<value>'.
func MakeEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// CheckerEnvs outputs envs for a given checker of service.
// It builds in this order:
// 1. Envs from Defaults.
// 2. Envs from Service.
// 3. Envs from Service.Checker.
func CheckerEnvs(g *ServiceDefaultsConfig, s *ServiceConfig) []string {
	return append(srvEnvs(g, s), s.Checker.Envs...)
}

// UpdaterEnvs outputs envs for a given updater of service.
// It builds in this order:
// 1. Envs from Defaults.
// 2. Envs from Service.
// 3. Envs from Service.Updater.
// 4. Add SKYUPD_TO_VERSION env.
func UpdaterEnvs(g *ServiceDefaultsConfig, s *ServiceConfig, toVersion string) []string {
	envs := append(srvEnvs(g, s), s.Updater.Envs...)
	if toVersion != "" {
		envs = append(envs, MakeEnv(EnvToVersion, toVersion))
	}
	return envs
}

func srvEnvs(g *ServiceDefaultsConfig, s *ServiceConfig) []string {
	envs := append(os.Environ(), g.Envs...)
	if s.Repo != "" {
		envs = append(envs, MakeEnv(EnvRepo, s.Repo))
	}
	if s.MainBranch != "" {
		envs = append(envs, MakeEnv(EnvMainBranch, s.MainBranch))
	}
	if s.BinDir != "" {
		envs = append(envs, MakeEnv(EnvBinDir, s.BinDir))
	}
	if s.MainProcess != "" {
		envs = append(envs, MakeEnv(EnvMainProcess, s.MainProcess))
	}
	return envs
}
