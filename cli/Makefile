.PHONY: help setup test test-cov lint clean format check build publish dev nixopus

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Setup Python environment and install dependencies
	@if command -v poetry >/dev/null 2>&1; then \
		echo "Poetry found. Installing dependencies..."; \
		poetry install --with dev --quiet; \
		echo "Environment ready! Use: make nixopus ARGS=\"command\""; \
	else \
		echo "Poetry not found. Installing Poetry..."; \
		curl -sSL https://install.python-poetry.org | python3 - >/dev/null 2>&1; \
		echo "Poetry installed. Please restart your shell or run: source ~/.bashrc (or ~/.zshrc)"; \
		echo "Then run 'make setup' again to install dependencies."; \
	fi

test: ## Run tests
	@poetry run pytest

test-cov: ## Run tests with coverage
	@poetry run pytest --cov=app --cov-report=term-missing --cov-report=html

lint: ## Run linting
	@poetry run flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics
	@poetry run flake8 . --count --exit-zero --max-complexity=10 --max-line-length=127 --statistics

format: ## Format code
	@poetry run black . --quiet
	@poetry run isort . --quiet

check: ## Run linting and tests
	$(MAKE) lint && $(MAKE) test

clean: ## Clean build artifacts
	@rm -rf build/ dist/ *.egg-info/ .pytest_cache/ htmlcov/ .coverage
	@find . -type d -name __pycache__ -delete
	@find . -type f -name "*.pyc" -delete

build: ## Build the package
	@poetry build

publish: ## Publish to PyPI
	@poetry publish

dev: ## Activate development shell
	@poetry shell


# -----------------------------------------------------------------------------
# Nixopus test CLI commands
# -----------------------------------------------------------------------------
nixopus: ## Run nixopus CLI
	@if [ -z "$(ARGS)" ]; then \
		poetry run nixopus --help; \
	else \
		poetry run nixopus $(ARGS); \
	fi

version: ## Show version
	@poetry run nixopus version

run: ## Run nixopus CLI directly
	poetry run nixopus

generate-docs: ## Generate CLI documentation
	typer app.main utils docs --output ../docs/cli/cli-reference.md --name nixopus
