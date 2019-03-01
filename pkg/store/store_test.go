package store

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewJSON(t *testing.T) {
	type service struct {
		Name   string
		Update Update
	}
	const srvCount = 100

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	defer func() {
		require.NoError(t, os.Remove(f.Name()))
	}()
	j, err := NewJSON(f.Name())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, j.Close())
	}()
	services := make([]service, srvCount)
	for i := 0; i < srvCount; i++ {
		services[i] = service{
			Name: fmt.Sprintf("service_%d", i),
			Update: Update{
				Tag: fmt.Sprintf("v1.%d", i),
				Timestamp: time.Now().
					Add(-time.Hour * time.Duration(24*rand.Intn(500))).
					UnixNano(),
			},
		}
	}
	for _, srv := range services {
		j.SetServiceLastUpdate(srv.Name, srv.Update)
	}
	for i, srv := range services {
		require.Equal(t, srv.Update, j.ServiceLastUpdate(srv.Name), i)
	}
}
