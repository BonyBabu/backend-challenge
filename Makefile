internal/generated/:
	docker run --rm \
	  -v ${PWD}:/local openapitools/openapi-generator-cli generate \
	  -i /local/api/openapi.yaml \
	  -g go-server \
	  -o /local/internal/generated/ \
	  --additional-properties=outputAsLibrary=true,sourceFolder=openapi

generate: internal/generated/

clean:
	rm -rf internal/generated/

.PHONY: run
run: generate
	go run cmd/foodorder/main.go

.PHONY: test
test:
	go generate ./... && go test ./...		
