package active

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/watercompany/skywire-updater/pkg/checker"
	"github.com/watercompany/skywire-updater/pkg/config"
	"github.com/watercompany/skywire-updater/pkg/logger"
	"github.com/watercompany/skywire-updater/pkg/store/services"
)

type git struct {
	// URL should be in the format /:owner/:Repository
	notifyURL string
	url       string
	service   string
	tag       string
	date      *time.Time
	exit      chan int
	log       *logger.Logger
	config.CustomLock
}

// newGit returns a new git fetcher
func newGit(service, url string, notifyURL string, log *logger.Logger) *git {
	retrievedStatus := services.GetStore().Get(service)
	log.Infof("retrieved status %+v", retrievedStatus)
	date := retrievedStatus.LastUpdated.Time

	return &git{
		url:       "https://api.github.com/repos" + url,
		notifyURL: notifyURL,
		tag:       "0.0.0",
		date:      &date,
		exit:      make(chan int),
		service:   service,
		log:       log,
	}
}

func (g *git) SetLastRelease(tag string, date *time.Time) {
	g.tag = tag

	if date != nil {
		g.date = date
	}
}

func (g *git) Check() error {
	return g.checkIfNew()
}

// ReleaseJSON encodes a release object in github to fetch its information
type ReleaseJSON struct {
	URL         string `json:"URL"`
	Name        string `json:"Name"` // Name encodes the name of the release, or its version
	PublishedAt string `json:"published_at"`
}

func (g *git) checkIfNew() error {
	if g.IsLock() {
		g.log.Warnf("service %s is already being updated... waiting for it to finish", g.service)
	}
	g.Lock()
	defer g.Unlock()

	release := g.fetchGithubRelease()
	publishedTime := parsePublishedTime(release, g.log)

	if g.date.Before(publishedTime) {
		g.log.Info("new version: ", release.URL, ". Published at: ", release.PublishedAt)

		// request notify url
		err := checker.NotifyUpdate(g.notifyURL, g.service, g.tag, g.tag, "token")
		if err != nil {
			return err
		}
		g.storeLastUpdated(publishedTime)
	} else {
		return ErrNoNewVersion
	}
	return nil
}

func (g *git) fetchGithubRelease() *ReleaseJSON {
	resp, err := http.Get(g.url + "/releases/latest")
	if err != nil {
		g.log.Fatal("cannot contact api, err ", err)
	}
	defer resp.Body.Close()
	release := &ReleaseJSON{}
	err = json.NewDecoder(resp.Body).Decode(release)
	if err != nil {
		g.log.Fatal("cannot unmarshal to a release object, err: ", err)
	}
	if release.PublishedAt == "" {
		g.log.Fatalf("unable to retrieve published at information from %s",
			g.url+"/release/latest. Make sure that the configuration repository exists")
	}

	return release
}

func parsePublishedTime(release *ReleaseJSON, log *logger.Logger) time.Time {
	publishedTime, err := time.Parse(time.RFC3339, release.PublishedAt)
	if err != nil {
		log.Fatal("cannot parse git release date: ", release.PublishedAt, " err: ", err)
	}
	return publishedTime
}

/*func (g *git) tryUpdate(release *ReleaseJSON) error {
	for i := 0; i < g.retries; i++ {
		err := <-g.updater.Update(g.service, release.Name, g.log)
		if err != nil {
			g.log.Errorf("error on update %s", err)

			if i == (g.retries - 1) {
				return fmt.Errorf("maximum retries attempted, service %s failed to update", g.service)
			} else {
				g.log.Infof("retry again in %s", g.retryTime.String())
			}
		} else {
			break
		}

		time.Sleep(g.retryTime)
	}
	return nil
}*/

func (g *git) storeLastUpdated(publishedTime time.Time) {
	g.date = &publishedTime
	storeService := services.Service{
		Name:        g.service,
		LastUpdated: services.NewTimeJSON(publishedTime),
	}
	services.GetStore().Store(&storeService)
}
