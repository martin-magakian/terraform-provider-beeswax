default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: build
build:
	go build -o bin/terraform-provider-beeswax

.PHONY: doc
doc:
	go generate ./...
