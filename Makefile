postgres:
	docker rm -f postgres
	docker run -p 5432:5432 -d --name postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret postgres:16-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

sleep10:
	sleep 10

local: redis postgres sleep10 createdb server

migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

dropdb:
	docker exec -it postgres dropdb simple_bank

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen --package mockwk --build_flags=--mod=mod --destination worker/mock/distributor.go github.com/SaishNaik/simplebank/worker TaskDistributor
	mockgen --package mockdb --build_flags=--mod=mod --destination db/mock/store.go github.com/SaishNaik/simplebank/db/sqlc Store


db_docs:
	 dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

new_migration:
		migrate create -ext sql -dir db/migrations -seq $(name)

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative\
  	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker rm -f redis
	docker run --name redis -p 6379:6379 -d redis:7.4-alpine

.PHONY: createdb postgres createdb migrateup migratedown migrateup1 migratedown1 dropdb sqlc test server mock db_docs db_schema proto evans local sleep10 new_migration

#start: postgres createdb migrateup


