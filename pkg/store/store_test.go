package store

//const jsonFile = `{
//  "skycoin": {
//	"name": "skycoin",
//    "last_updated": "2018-04-26T02:06:00Z"
//  },
//  "skywire": {
//	"name": "skywire",
//    "last_updated": "2018-06-26T02:06:00Z"
//  }
//}`
//
//func TestDecodeFile(t *testing.T) {
//	assert.NotNil(t, decodeServicesMap([]byte(jsonFile)))
//}
//
//func TestGet(t *testing.T) {
//	const (
//		serviceName        = "skycoin"
//		serviceLastUpdated = "2018-04-26T02:06:00Z"
//	)
//	jsonStore := JSONServices{
//		services: decodeServicesMap([]byte(jsonFile)),
//	}
//	service := jsonStore.Get(serviceName)
//	assert.Equal(t, service.LastUpdated.Format(time.RFC3339), serviceLastUpdated)
//}
