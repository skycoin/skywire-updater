package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/watercompany/skywire-updater/pkg/store"
)

// CheckerType represents a Checker's type.
type CheckerType string

const (
	// GithubReleaseCheckerType type.
	GithubReleaseCheckerType = CheckerType("github-release")

	// ScriptCheckerType type.
	ScriptCheckerType = CheckerType("script")
)

var checkerTypes = []CheckerType{
	GithubReleaseCheckerType,
	ScriptCheckerType,
}

// Release is obtained from a check.
type Release struct {
	HasUpdate   bool            `json:"update_available"`
	Version     string          `json:"release_version"`
	Timestamp   time.Time       `json:"release_timestamp"`
	CheckerType CheckerType     `json:"checker_type"`
	GitRelease  *GitReleaseBody `json:"git_release,omitempty"`
}

// Checker represents a Checker implementation.
type Checker interface {
	Check(ctx context.Context) (*Release, error)
}

// NewChecker creates a new Checker and panics on failure.
func NewChecker(db store.Store, srvName string, c ServiceConfig, d *DefaultsConfig) Checker {
	switch c.Checker.Type {
	case GithubReleaseCheckerType:
		return NewGithubReleaseChecker(db, srvName, c)
	case ScriptCheckerType:
		return NewScriptChecker(srvName, c, d)
	default:
		log.Fatalf("invalid checker type '%s' at 'services[%s].checker.type' when expecting: %v",
			c.Checker.Type, srvName, checkerTypes)
		return nil
	}
}

// ScriptChecker checks via scripts.
type ScriptChecker struct {
	c   ServiceConfig
	d   *DefaultsConfig
	log *logging.Logger
}

// NewScriptChecker uses a given script as a checker.
func NewScriptChecker(srvName string, c ServiceConfig, d *DefaultsConfig) *ScriptChecker {
	return &ScriptChecker{
		c:   c,
		d:   d,
		log: logging.MustGetLogger("script-checker." + srvName),
	}
}

// Check checks for updates.
func (sc *ScriptChecker) Check(ctx context.Context) (*Release, error) {
	check := sc.c.Checker
	cmd := exec.Command(check.Interpreter, append([]string{check.Script}, check.Args...)...) //nolint:gosec
	cmd.Env = CheckerEnvs(sc.d, &sc.c)
	hasUpdate, err := ExecuteScript(ctx, sc.log, cmd)
	if err != nil {
		return nil, err
	}
	return &Release{
		HasUpdate:   hasUpdate,
		Timestamp:   time.Now(),
		CheckerType: ScriptCheckerType,
	}, nil
}

// GithubReleaseChecker checks for available updates via the github API.
type GithubReleaseChecker struct {
	srvName string
	c       ServiceConfig
	db      store.Store
	log     *logging.Logger
}

// NewGithubReleaseChecker creates a new GithubReleaseChecker.
func NewGithubReleaseChecker(db store.Store, srvName string, c ServiceConfig) *GithubReleaseChecker {
	return &GithubReleaseChecker{
		srvName: srvName,
		c:       c,
		db:      db,
		log:     logging.MustGetLogger("release-checker." + srvName),
	}
}

// Check checks for updates.
func (gc *GithubReleaseChecker) Check(ctx context.Context) (*Release, error) {
	body, err := gc.fetchFromGit(ctx)
	if err != nil {
		return nil, err
	}
	pubAt, err := body.ParsePubAt()
	if err != nil {
		return nil, err
	}
	last := gc.db.ServiceLastUpdate(gc.srvName)
	hasUpdate := last.IsEmpty() || last.Timestamp < pubAt.UnixNano()
	return &Release{
		HasUpdate:   hasUpdate,
		Version:     body.TagName,
		Timestamp:   pubAt,
		CheckerType: GithubReleaseCheckerType,
		GitRelease:  body,
	}, nil
}

// GitReleaseBody is the response body of the github API call.
type GitReleaseBody struct {
	URL     string `json:"url,omitempty"`
	TagName string `json:"tag_name,omitempty"`
	PubAt   string `json:"published_at,omitempty"`
	Body    string `json:"body,omitempty"`
}

// ParsePubAt parses the published_at field.
func (grb GitReleaseBody) ParsePubAt() (time.Time, error) {
	return time.Parse(time.RFC3339, grb.PubAt)
}

func (gc *GithubReleaseChecker) fetchFromGit(ctx context.Context) (*GitReleaseBody, error) {
	repo := strings.TrimPrefix(gc.c.Repo, "github.com")
	url := "https://" + path.Join("api.github.com/repos/", repo, "/releases/latest")
	gc.log.Infoln("Request URL:", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if ok := addBasicAuth(req); ok {
		gc.log.Infof("Added basic auth headers.")
	}
	if err != nil {
		gc.log.WithError(err).Fatalln("failed to formulate 'fetchFromGit' request")
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body GitReleaseBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		gc.log.WithError(err).Fatalln("unrecognised json body")
		return nil, err
	}

	jStr, _ := json.MarshalIndent(body, "", "    ") //nolint:errcheck
	gc.log.Infoln(fmt.Sprintf("Response (%s):", resp.Status), string(jStr))
	return &body, nil
}

func addBasicAuth(req *http.Request) bool {
	usr, usrOK := os.LookupEnv(EnvGithubUsername)
	pac, pacOK := os.LookupEnv(EnvGithubAccessToken)
	if usrOK && pacOK {
		req.SetBasicAuth(usr, pac)
		return true
	}
	return false
}
