# Verdagostack Makefile
# ──────────────────────────────────────────────────────────────────────
#
# License headers: Apache 2.0 (addlicense, Go sources only; ignores vendor)
ADDLICENSE   := go run github.com/google/addlicense@v1.2.0
LFLAGS       := -c "Verda Cloud Oy" -l apache -y 2026
GOSRC_FIND   = find . -name '*.go' -not -path './vendor/*'

.PHONY: build test lint fmt license license-check security setup pre-commit
.PHONY: tag tag-list tag-delete release changelog help

# ─── Development Setup ──────────────────────────────────────────────

setup: ## Set up development environment (tools + hooks)
	@./scripts/setup.sh

# ─── Build & Test ────────────────────────────────────────────────────

build: ## Build all packages
	@echo "→ Building..."
	@go build ./...
	@echo "✓ Build successful!"

test: ## Run all tests
	@echo "→ Running tests..."
	@go test ./... -count=1
	@echo "✓ Tests passed!"

# ─── Code Quality ───────────────────────────────────────────────────

lint: ## Run golangci-lint
	@echo "→ Running golangci-lint..."
	@golangci-lint run ./...
	@echo "✓ Linting complete!"

fmt: ## Format Go code
	@echo "→ Formatting Go code..."
	@gofmt -w -s .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "  ℹ goimports not found, skipping import formatting"; \
	fi
	@echo "✓ Formatting complete!"

license: ## Add Apache 2.0 license headers to all Go source files
	@echo "→ Adding license headers to .go files..."
	@$(GOSRC_FIND) -exec $(ADDLICENSE) $(LFLAGS) -v {} +
	@echo "✓ License headers applied!"

license-check: ## Verify all Go files have the Apache 2.0 license header
	@echo "→ Checking license headers on .go files..."
	@$(GOSRC_FIND) -exec $(ADDLICENSE) -check $(LFLAGS) {} +
	@echo "✓ License headers OK!"

security: ## Run security checks (gosec + govulncheck + gitleaks + trivy + osv-scanner)
	@./scripts/security-scan.sh

pre-commit: ## Run all pre-commit hooks on all files
	@echo "→ Running pre-commit hooks on all files..."
	@pre-commit run --all-files

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
