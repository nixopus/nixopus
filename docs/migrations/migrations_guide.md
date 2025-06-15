# 📚 Database Migrations Guide – Nixopus

This guide explains how database migrations work in Nixopus, how to create and apply them, and best practices for using them effectively.

---

## 🧰 Tool We Use: `golang-migrate`

Nixopus uses [`golang-migrate`](https://github.com/golang-migrate/migrate) to manage SQL-based database migrations in Go. This tool supports programmatic and CLI-based migration execution, and we’ve embedded it into our application lifecycle so that migrations run automatically on startup.

---

## 📂 Where Migrations Live

All migrations are stored in the `migrations/` directory (relative to the module where migration logic exists). Each migration consists of two SQL files:

```
<version>_<name>.up.sql   # Migration to apply
<version>_<name>.down.sql # Migration to rollback
```

Example:

```
000001_create_users_table.up.sql
000001_create_users_table.down.sql
```

---

## ✨ How Migrations Are Applied

On application startup, the code in `migration.go` runs all unapplied `.up.sql` files in order, using `golang-migrate`'s `Up()` function.

**Automatic Execution:**
- Migrations run when the app starts.
- Already-applied migrations are skipped.
- Locking ensures only one process can apply them at a time (via advisory locks).

> ✅ No manual steps are needed if your migration files are correctly placed and committed.

---

## ✏️ Creating a New Migration

Use the CLI to generate a new migration file pair:

```bash
migrate create -ext sql -dir ./migrations -seq add_new_column_to_users
```

This creates:

```
000002_add_new_column_to_users.up.sql
000002_add_new_column_to_users.down.sql
```

### ✍️ Edit the files:

**Up (apply):**

```sql
ALTER TABLE users ADD COLUMN nickname TEXT;
```

**Down (rollback):**

```sql
ALTER TABLE users DROP COLUMN nickname;
```

---

## 🕒 When Migrations Run

- Automatically on service startup.
- On local dev, you can trigger them manually via CLI (see below).
- Only `.up.sql` files are executed in forward direction, and only if not already applied.

---

## 🔁 Rollbacks

To undo one or more migrations:

### Using CLI:

```bash
migrate -path ./migrations -database <DB_URL> down 1     # Rollback last 1 migration
migrate -path ./migrations -database <DB_URL> down        # Rollback all
```

### Programmatically:

In Go:

```go
_ = m.Steps(-1)   // Rollback one
_ = m.Down()      // Rollback all
```

### Dirty state recovery:

If a migration fails midway:
- The DB may be marked “dirty”
- Use `migrate force <version>` to fix state before rerunning

---

## 🧪 How to Test Migrations

1. Spin up a local DB (Docker or Postgres locally)
2. Run all migrations up:
   ```bash
   migrate -path ./migrations -database <DB_URL> up
   ```
3. Run all down:
   ```bash
   migrate -path ./migrations -database <DB_URL> down
   ```
4. Run tests (if defined) to verify schema is as expected

---

## 🧾 Naming Conventions

- Use `snake_case` for clarity
- Be descriptive: `add_index_to_org_table`, `create_user_settings_table`, etc.
- Keep versions sequential (`000001`, `000002`, ...) or use timestamps (`20250615120000_create_table.up.sql`)

---

## 🛠️ Best Practices

- Keep migrations **small and focused**
- Always add both `.up.sql` and `.down.sql`
- **Test migrations** in a clean DB locally before pushing
- Do **not edit** already-applied migration files—create a new one
- Include migration tests in CI/CD
- Prefer **embedded migrations** in binary for consistent deploys

---

## 📎 Helpful Commands

```bash
# Install CLI
brew install golang-migrate

# Run migrations
migrate -path ./migrations -database postgres://user:pass@localhost:5432/db up

# Rollback
migrate -path ./migrations -database postgres://user:pass@localhost:5432/db down 1

# Fix dirty state
migrate -path ./migrations -database <DB_URL> force <version>
```

---

## 📌 Example Migration

### `000003_add_status_to_orders.up.sql`

```sql
ALTER TABLE orders ADD COLUMN status VARCHAR(20) DEFAULT 'pending';
```

### `000003_add_status_to_orders.down.sql`

```sql
ALTER TABLE orders DROP COLUMN status;
```

---

## 🙋 Common Issues

| Issue | Fix |
|------|-----|
| Migration failed and DB is "dirty" | Run `migrate force <version>` |
| "No change" message | All migrations already applied |
| Rollback fails | Ensure `.down.sql` is defined and reversible |
| Duplicate version error | Check your filenames and sequencing |

---

## 📫 Questions?

Reach out in the #backend Slack channel or ping `@<migration-maintainer>` if you're stuck!