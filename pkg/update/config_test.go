package update

//func TestServices(t *testing.T) {
//	var expectedServiceMaps = map[string]ServiceConfig{
//		"skywire-manager": {
//			Process: "skywire-manager",
//			Repo:    "github.com/skycoin/skywire",
//			Checker: CheckerConfig{
//				Type:        ScriptCheckerType,
//				Interpreter: "/bin/bash",
//				Script:      "generic-service-check-update.sh",
//			},
//			Updater: UpdaterConfig{
//				Type:        ScriptUpdaterType,
//				Interpreter: "/bin/bash",
//				Script:      "skywire.sh",
//				Args:        []string{"-web-dir ${GOPATH}/pkg/github.com/skycoin/skywire/static/skywire-manager"},
//			},
//		},
//		"skywire-node": {
//			Process: "skywire-node",
//			Repo:    "github.com/skycoin/skywire",
//			Checker: CheckerConfig{
//				Type:        ScriptCheckerType,
//				Interpreter: "/bin/bash",
//				Script:      "generic-service-check-update.sh",
//			},
//			Updater: UpdaterConfig{
//				Type:        ScriptUpdaterType,
//				Interpreter: "/bin/bash",
//				Script:      "skywire.sh",
//				Args: []string{
//					"-connect-manager -manager-address 127.0.0.1:5998",
//					"-manager-web 127.0.0.1:8000",
//					"-discovery-address discovery.skycoin.net:5999-034b1cd4ebad163e457fb805b3ba43779958bba49f2c5e1e8b062482904bacdb68",
//					"-address :5000",
//					"-web-port :6001",
//				},
//			},
//		},
//	}
//
//	c := NewConfig("../../configuration.skywire.yml")
//
//	assert.Equal(t, expectedServiceMaps, c.Services)
//}
