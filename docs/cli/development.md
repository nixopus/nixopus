# CLI Development

Basic development guide for contributing to the Nixopus CLI.

## Project Structure

```
cli/
├── app/                 # Main application package
│   ├── main.py          # CLI entry point
│   ├── commands/        # Command implementations
│   └── utils/           # Shared utilities
├── pyproject.toml       # Project configuration (Poetry)
├── poetry.lock          # Dependency lock file
└── Makefile             # Development commands
```

## Development Setup

### Prerequisites
- Python 3.9+
- Poetry (for dependency management)
- Git

### Installation

```bash
# Clone and navigate
git clone https://github.com/raghavyuva/nixopus.git
cd nixopus/cli

# Install dependencies
poetry install --with dev

# Activate virtual environment
poetry shell

# Install CLI in development mode
pip install -e .

# Verify installation
nixopus --help
```

## Testing

```bash
# Set development environment (required for tests)
export ENV=DEVELOPMENT

# Run tests
make test

# Run with coverage
make test-cov

# Run specific test
poetry run pytest tests/test_commands_version.py
```

## Available Make Commands

```bash
make help          # Show available commands
make install       # Install dependencies
make test          # Run test suite
make test-cov      # Run tests with coverage
make lint          # Run code linting
make format        # Format code
make clean         # Clean build artifacts
make build         # Build distribution
```

## Contributing

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature
   ```

2. **Make changes and test**
   ```bash
   export ENV=DEVELOPMENT
   make test
   ```

3. **Commit and submit pull request**
   ```bash
   git add .
   git commit -m "Description of changes"
   git push origin feature/your-feature
   ```

## Dependencies

### Core Dependencies
- **typer**: CLI framework
- **rich**: Terminal formatting  
- **pydantic**: Data validation
- **requests**: HTTP library
- **pyyaml**: YAML parsing

### Development Dependencies
- **pytest**: Testing framework
- **pytest-cov**: Coverage reporting
- **black**: Code formatting
- **flake8**: Code linting