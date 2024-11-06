postgres:
	docker run --network bank-network --rm -p 5432:5432 -d --name postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret postgres:16-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

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
	go test -v -cover ./...

server:
	go run main.go

mock:
	 mockgen --package mockdb --build_flags=--mod=mod --destination db/mock/store.go github.com/SaishNaik/simplebank/db/sqlc Store

.PHONY: createdb postgres createdb migrateup migratedown migrateup1 migratedown1 dropdb sqlc test server mock

#start: postgres createdb migrateup


