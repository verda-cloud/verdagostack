# Verdagostack Makefile
# ──────────────────────────────────────────────────────────────────────

.PHONY: build test vet lint tag tag-list tag-delete release changelog help

# ─── Build & Test ────────────────────────────────────────────────────

build: ## Build all packages
	go build ./...

test: ## Run all tests
	go test ./... -count=1

vet: ## Run go vet
	go vet ./...

lint: vet ## Alias for vet (add staticcheck/golangci-lint here if desired)

# ─── Version Tagging ─────────────────────────────────────────────────
#
# Usage:
#   make tag VERSION=v0.2.0                  # create and push a new tag
#   make tag VERSION=v0.2.0 MSG="my note"    # with a custom annotation message
#   make tag-list                             # list all version tags
#   make tag-delete VERSION=v0.2.0           # delete a tag locally and remotely

VERSION  ?=
MSG      ?=

tag: ## Create an annotated version tag and push it (VERSION required)
ifndef VERSION
	$(error VERSION is required. Usage: make tag VERSION=v0.2.0)
endif
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "Error: VERSION must be valid semver (e.g. v0.2.0)"; exit 1; \
	fi
	@if git rev-parse "$(VERSION)" >/dev/null 2>&1; then \
		echo "Error: tag $(VERSION) already exists"; exit 1; \
	fi
	$(eval TAG_MSG := $(if $(MSG),$(MSG),$(VERSION)))
	git tag -a "$(VERSION)" -m "$(TAG_MSG)"
	git push origin "$(VERSION)"
	@echo ""
	@echo "✓ Tag $(VERSION) created and pushed."
	@echo "  Downstream: go get github.com/verda-cloud/verdagostack@$(VERSION)"

tag-list: ## List all version tags (newest first)
	@git tag -l 'v*' --sort=-v:refname | head -20
	@echo ""
	@echo "Latest: $$(git tag -l 'v*' --sort=-v:refname | head -1)"

tag-delete: ## Delete a version tag locally and on remote (VERSION required)
ifndef VERSION
	$(error VERSION is required. Usage: make tag-delete VERSION=v0.2.0)
endif
	git tag -d "$(VERSION)"
	git push origin --delete "$(VERSION)"
	@echo "✓ Tag $(VERSION) deleted."

# ─── Release Management ─────────────────────────────────────────────
#
# Usage:
#   make release VERSION=v0.2.0   # generate changelog and commit release
#   make changelog                # preview unreleased changelog entries

release: ## Prepare a new release: generate changelog and tag (VERSION required)
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v0.2.0)
endif
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "Error: VERSION must be valid semver (e.g. v0.2.0)"; exit 1; \
	fi
	@if ! command -v git-cliff >/dev/null 2>&1; then \
		echo "Error: git-cliff is not installed"; \
		echo "Install: cargo install git-cliff  OR  brew install git-cliff"; \
		exit 1; \
	fi
	@echo "→ Generating CHANGELOG.md with git-cliff..."
	@git-cliff --tag $(VERSION) -o CHANGELOG.md
	@echo "✓ Updated CHANGELOG.md"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Review CHANGELOG.md"
	@echo "  2. git add CHANGELOG.md && git commit -m 'chore(release): prepare $(VERSION)'"
	@echo "  3. make tag VERSION=$(VERSION)"

changelog: ## Preview unreleased changelog entries (requires git-cliff)
	@if ! command -v git-cliff >/dev/null 2>&1; then \
		echo "Error: git-cliff is not installed"; \
		echo "Install: cargo install git-cliff  OR  brew install git-cliff"; \
		exit 1; \
	fi
	@git-cliff --unreleased

# ─── Help ────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
