package mockstore

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/watercompany/skywire-services/pkg/transport-discovery/store/tests"
)

func TestSQL(t *testing.T) {
	suite.Run(t, &tests.TransportSuite{Store: NewStore()})
}
