.PHONY: test
test:
	./scripts/test.sh

.PHONY: dep
dep:
	go mod download

.PHONY: vet
vet:
	go vet

.PHONY: gen
gen:
	./scripts/generate.sh
	./scripts/format.sh

.PHONY: lint
lint:
	./scripts/lint.sh

.PHONY: format
format:
	./scripts/format.sh
