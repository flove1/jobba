include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the src/api application
.PHONY: run
run:
	@go run ./src/api -db-dsn=${JOBBE_DB_DSN} -smtp-host="smtp.office365.com" -smtp-username=${EMAIL} -smtp-password=${EMAIL_PASSWORD} -smtp-sender="Jobbe <${EMAIL}>"

## db/psql: connect to the database using psql
.PHONY: psql
psql:
	psql ${JOBBE_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up/all
db/migrations/up/all: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${JOBBE_DB_DSN} up

## db/migrations/up: apply one up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${JOBBE_DB_DSN} up 1

## db/migrations/up: apply all down database migrations
.PHONY: db/migrations/down/all
db/migrations/down/all: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${JOBBE_DB_DSN} down

## db/migrations/up: apply one down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${JOBBE_DB_DSN} down 1

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## audit: tidy abd verify dependencies
PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor