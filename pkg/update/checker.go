package update

import (
	"context"
	"encoding/json"
	"net/http"
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
	GithubReleaseCheckerType = CheckerType("github_release")

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
func NewChecker(log *logging.Logger, db store.Store, srvName string, srvConfig ServiceConfig) Checker {
	switch srvConfig.Checker.Type {
	case GithubReleaseCheckerType:
		return NewGithubReleaseChecker(log, db, srvName, srvConfig)
	case ScriptCheckerType:
		return NewScriptChecker(log, srvName, srvConfig)
	default:
		log.Fatalf("invalid checker type '%s' at 'services[%s].checker.type' when expecting: %v",
			srvConfig.Checker.Type, srvName, checkerTypes)
		return nil
	}
}

// ScriptChecker checks via scripts.
type ScriptChecker struct {
	srvName string
	c       ServiceConfig
	log     *logging.Logger
}

// NewScriptChecker ...
func NewScriptChecker(log *logging.Logger, srvName string, c ServiceConfig) *ScriptChecker {
	return &ScriptChecker{
		srvName: srvName,
		c:       c,
		log:     log,
	}
}

// Check ...
func (sc *ScriptChecker) Check(ctx context.Context) (*Release, error) {
	check := sc.c.Checker
	cmd := exec.Command(check.Interpreter, append([]string{check.Script}, check.Args...)...) //nolint:gosec
	cmd.Env = sc.c.CheckerEnvs()
	hasUpdate, err := executeScript(ctx, cmd, sc.log)
	if err != nil {
		return nil, err
	}
	return &Release{
		HasUpdate:   hasUpdate,
		Timestamp:   time.Now(),
		CheckerType: ScriptCheckerType,
	}, nil
}

// GithubReleaseChecker ...
type GithubReleaseChecker struct {
	srvName string
	sc      ServiceConfig
	db      store.Store
	log     *logging.Logger
}

// NewGithubReleaseChecker ...
func NewGithubReleaseChecker(log *logging.Logger, db store.Store, srvName string, sc ServiceConfig) *GithubReleaseChecker {
	return &GithubReleaseChecker{
		srvName: srvName,
		sc:      sc,
		db:      db,
		log:     log,
	}
}

// Check ...
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

// GitReleaseBody ...
type GitReleaseBody struct {
	URL     string `json:"url,omitempty"`
	TagName string `json:"tag_name,omitempty"`
	PubAt   string `json:"published_at,omitempty"`
	Body    string `json:"body,omitempty"`
}

// ParsePubAt ...
func (grb GitReleaseBody) ParsePubAt() (time.Time, error) {
	return time.Parse(time.RFC3339, grb.PubAt)
}

func (gc *GithubReleaseChecker) fetchFromGit(ctx context.Context) (*GitReleaseBody, error) {
	repo := strings.TrimPrefix(gc.sc.Repo, "github.com")
	url := "https://" + path.Join("api.github.com/repos/", repo, "/releases/latest")
	gc.log.Infoln("Request URL:", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
	return &body, nil
}
