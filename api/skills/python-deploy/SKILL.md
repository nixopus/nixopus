---
name: python-deploy
description: Build and deploy Python applications — uv, poetry, pdm, pipenv, pip, framework detection, and Dockerfile patterns. Use when deploying a Python project, or when requirements.txt, pyproject.toml, or Pipfile is detected.
metadata:
  version: "1.0"
---

# Python Deployment

## Detection

Project is Python if any of these exist:

| File | Package manager / style |
|---|---|
| `pyproject.toml` | Modern (PEP 517/518), uv, poetry, pdm, or pip |
| `requirements.txt` | pip |
| `setup.py` | setuptools |
| `Pipfile` | Pipenv |

Entry point files (checked in order): `app.py`, `main.py`, `server.py`, `run.py`, `wsgi.py`, `asgi.py`

## Versions

Python version priority:

1. `.python-version` file or `mise.toml` / `.tool-versions`
2. `pyproject.toml` → `[project].requires-python` (e.g. `>=3.11`)
3. `Pipfile` → `[requires].python_version`
4. `runtime.txt`
5. Defaults to **3.13**

## Package Managers

Detect in order:

1. **Lock files:** `poetry.lock` → Poetry, `pdm.lock` → PDM, `Pipfile.lock` → Pipenv, `uv.lock` → uv
2. **Manifest:** `pyproject.toml` (poetry/pdm/uv) or `Pipfile` (pipenv) or `requirements.txt` (pip)
3. **Default:** pip with `requirements.txt`

### Install Commands

| Manager | Install command |
|---|---|
| pip | `pip install --no-cache-dir -r requirements.txt` |
| Poetry | `poetry install --only main --no-interaction` |
| PDM | `pdm install --prod --no-lock` |
| uv | `uv sync --frozen --no-dev` |
| Pipenv | `pipenv install --deploy --system` |

## Runtime Variables

Configure the Python runtime for production: enable the fault handler for crash diagnostics (`PYTHONFAULTHANDLER=1`), disable output buffering for real-time logs (`PYTHONUNBUFFERED=1`), randomize hash seeds (`PYTHONHASHSEED=random`), and skip bytecode generation (`PYTHONDONTWRITEBYTECODE=1`). Suppress pip version checks with `PIP_DISABLE_PIP_VERSION_CHECK=1`.

## Build & Start

### Start Command Resolution

1. Framework-specific (see below)
2. `pyproject.toml` → `[project.scripts]` or `[tool.poetry.scripts]`
3. `setup.py` → `entry_points`
4. `manage.py` → Django
5. Root entry files (first found): `app.py`, `main.py`, `server.py`, `run.py` → `python <file>`
6. Conventions: `wsgi.py`, `asgi.py`

Most Python web apps have no build step (interpreted).

## Port Detection

1. `PORT` from `.env` or `.env.example`
2. Framework defaults:

| Framework | Default port |
|---|---|
| Flask | 5000 |
| Django / FastAPI / Starlette / Gunicorn / Uvicorn | 8000 |
| Streamlit | 8501 |
| Gradio | 7860 |
| Default | 8000 |

## Framework Detection

From dependencies in `pyproject.toml`, `requirements.txt`, or `Pipfile`:

| Package pattern | Framework | Category |
|---|---|---|
| `flask` | Flask | Backend |
| `django` | Django | FullStack |
| `fastapi` | FastAPI | Backend |
| `starlette` | Starlette | Backend |
| `python-fasthtml` | FastHTML | Backend |
| `uvicorn` | ASGI server | Backend |
| `gunicorn` | WSGI server | Backend |
| `streamlit` | Streamlit | App |
| `gradio` | Gradio | App |
| `pelican` | Pelican | Static |
| `mkdocs` | MkDocs | Docs |

### Framework-Specific Start Commands

**Flask** — `gunicorn <module>:app -b 0.0.0.0:${PORT:-5000}` if `gunicorn` is a dependency, else `flask run`
**Django** — `python manage.py migrate && gunicorn <wsgi_module>:application -b 0.0.0.0:${PORT:-8000} --workers 2`. Detect app from `WSGI_APPLICATION` in settings.
**FastAPI** — `uvicorn <module>:app --host 0.0.0.0 --port ${PORT:-8000} --workers 1`
**FastHTML** — `uvicorn <module>:app --host 0.0.0.0 --port ${PORT:-8000}`
**Streamlit** — `streamlit run <entry>.py --server.port ${PORT:-8501} --server.address 0.0.0.0`

## System Dependencies

| Package | Build-time | Runtime |
|---|---|---|
| `psycopg2` | `libpq-dev`, `gcc` | `libpq5` |
| `mysqlclient` | `default-libmysqlclient-dev`, `gcc` | `libmariadb3` |
| `pillow` | `libjpeg-dev`, `zlib1g-dev`, `libfreetype6-dev` | — |
| `cryptography` | `libssl-dev`, `libffi-dev` | — |
| `pycairo` | `libcairo2-dev`, `pkg-config` | `libcairo2` |
| `lxml` | `libxml2-dev`, `libxslt1-dev` | `libxml2`, `libxslt1.1` |
| `pycurl` | `libcurl4-openssl-dev` | — |
| `pdf2image` | — | `poppler-utils` |
| `pydub` | — | `ffmpeg` |
| `playwright` | — | Chromium headless shell |

## Install Stage Optimization

Copy `pyproject.toml` + lockfile first for layer caching, then full source.

## Caching

Leverage BuildKit cache mounts to speed up rebuilds:

- pip: `RUN --mount=type=cache,target=/root/.cache/pip pip install -r requirements.txt`
- uv: `RUN --mount=type=cache,target=/root/.cache/uv uv sync --frozen`

## Environment Variable Semantics

| Pattern | Likely dependency |
|---|---|
| `DATABASE_URL` | PostgreSQL, MySQL |
| `REDIS_URL` | Redis |
| `SECRET_KEY` | Django, Flask |
| `DJANGO_SETTINGS_MODULE` | Django |
| `AWS_*` / `S3_*` | AWS |

## Dockerfile Patterns

### uv (with lockfile)

```dockerfile
FROM python:3.13-slim AS base

FROM base AS deps
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN pip install uv && uv sync --frozen

FROM base AS runtime
WORKDIR /app
ENV PYTHONUNBUFFERED=1 PYTHONDONTWRITEBYTECODE=1
COPY --from=deps /app/.venv /app/.venv
ENV PATH="/app/.venv/bin:$PATH"
COPY . .
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### pip + requirements.txt

```dockerfile
FROM python:3.13-slim AS base

FROM base AS deps
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

FROM base AS runtime
WORKDIR /app
ENV PYTHONUNBUFFERED=1 PYTHONDONTWRITEBYTECODE=1
COPY --from=deps /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY . .
EXPOSE 8000
CMD ["gunicorn", "wsgi:app", "-b", "0.0.0.0:8000"]
```

### Django with collectstatic

```dockerfile
FROM python:3.13-slim AS base

FROM base AS deps
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

FROM base AS build
WORKDIR /app
COPY --from=deps /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY . .
ENV DJANGO_SETTINGS_MODULE=project.settings
RUN python manage.py collectstatic --noinput

FROM base AS runtime
WORKDIR /app
ENV PYTHONUNBUFFERED=1 PYTHONDONTWRITEBYTECODE=1
COPY --from=build /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=build /app/static /app/static
COPY --from=build /app /app
EXPOSE 8000
CMD ["gunicorn", "project.wsgi:application", "-b", "0.0.0.0:8000"]
```

## Gotchas

- `psycopg2` requires `libpq-dev` and `gcc` at build time — use `psycopg2-binary` instead to avoid native compilation
- uv `--frozen` fails if `uv.lock` doesn't exist — use `uv sync` without `--frozen` when no lockfile is present
- Django `collectstatic` needs `DJANGO_SETTINGS_MODULE` and may require a dummy `SECRET_KEY` at build time
- Poetry creates virtual environments outside the project by default — set `POETRY_VIRTUALENVS_IN_PROJECT=true` for Docker builds
- FastAPI/Uvicorn must bind to `0.0.0.0`, not `127.0.0.1`, to be reachable from outside the container
- Python `site-packages` path includes the minor version (e.g. `python3.13`) — the COPY path in multi-stage Dockerfiles must match the exact Python version
