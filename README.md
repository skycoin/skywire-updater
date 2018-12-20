# skywire-services
Certain services that are required for Skywire to function.

## Transport Discovery
The Transport Discovery is a service that exposes a RESTFUL interface and interacts with a database on the back-end.

### Running:

Setup the database:
```
docker run -p5432:5432 -d postgres
psql -h 127.0.0.1 --user postgres -c 'create database test'
```

Run the tests:
```
POSTGRES_TEST_DSN='user=postgres host=127.0.0.1 database=test sslmode=disable' go test ./pkg/transport-discovery/... -race
```

Start the server:
```
go run cmd/transport-discovery/main.go serve -bind :8081 -db 'user=postgres host=127.0.0.1 database=test sslmode=disable'
```

Send a request:
```
curl localhost:8081/ids/1 \
  -H Sw-Public:02aaeeedea55c1f216f863c0e750346fe2d0ac40b937a72d81b8460a7b136d8662 \
  -H Sw-Sig:d4bc234d3acf5b4643abb205328459cde22b94258f899052d1b457ee3be732ac0f80b26b1091504c0f887c7dd80a790949e880bc01b97184d2e6b20fb87f4c5e00 \
  -H Sw-Nonce:0 | jq
```
