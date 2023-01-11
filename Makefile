dev:
	goreleaser build --single-target --snapshot --rm-dist

test:
	sh -c 'cd ./pkg/cuedefs && cue vet ./tests/... -c'
	sh -c 'cd ./pkg/cuedefs && cue eval ./tests/... -c'
	go test ./... -race -count=1
	golangci-lint run

e2e:
	# go run . -test.v -test.run inmemory-async
	cd ./tests && go run . -test.v

snapshot:
	goreleaser release --snapshot --skip-publish --rm-dist

build-ui:
	cd ui && yarn
	cd ui && yarn build
	cp -r ./ui/dist/* ./pkg/devserver/static/

build:
	goreleaser build

gql:
	go run github.com/99designs/gqlgen --verbose --config ./pkg/coreapi/gqlgen.yml
