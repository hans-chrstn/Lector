# Lector

Lector is a neutral, open-source document management and reading application. It is designed to allow users to organize and read their own locally-stored files (such as EPUB, PDF, and CBZ/CBR) and to extend the application's functionality through a Lua-based plugin system.

---

## Security and Terms

Lector is a neutral tool. It does not provide, host, or distribute any content. Users are solely responsible for the legality of the content they import and the plugins they choose to install. For full details, see [DISCLAIMER.md](DISCLAIMER.md).

---

## Deployment

Lector is designed to be easily deployed via Docker. It runs as an unprivileged user for maximum security.

### Docker Environment Variables

| Variable             | Description                       | Default               |
| :------------------- | :-------------------------------- | :-------------------- |
| `AUTH_USER`          | Username for Basic Authentication | (None)                |
| `AUTH_PASSWORD`      | Password for Basic Authentication | (None)                |
| `MAX_UPLOAD_SIZE`    | Maximum upload size in MB         | `100`                 |
| `DB_DRIVER`          | Database driver (`sqlite` or `postgres`) | `sqlite` |
| `DATABASE_PATH`      | Path to SQLite file (SQLite only) | `/app/data/lector.db` |
| `DATABASE_URL`       | DSN/URL (Postgres only)           | (None)                |
| `CORS_ALLOW_ORIGINS` | Permitted origins for CORS        | (Same-origin only)    |
| `PORT`               | Port to listen on                 | `3000`                |

### PostgreSQL Support

Lector supports PostgreSQL for high-concurrency environments. To use it, set `DB_DRIVER=postgres` and provide a `DATABASE_URL` (e.g., `postgres://user:pass@host:5432/db`).

### Migrating from SQLite to PostgreSQL

If you are already using SQLite and want to move your data to PostgreSQL, use the built-in migration tool:

```bash
docker run --rm \
  -v lector_data:/app/data \
  -e SQLITE_PATH=/app/data/lector.db \
  -e DATABASE_URL=postgres://user:pass@host:5432/db \
  -e MIGRATE_DIRECTION=sqlite2postgres \
  lector ./db-migrator
```

### Building and Running

```bash
docker build -t lector .
docker run -p 3000:3000 \
  -e AUTH_USER=admin \
  -e AUTH_PASSWORD=yourpassword \
  -v lector_data:/app/data \
  lector
```

---

## Plugin API Documentation

Lector plugins use Lua to extend the application. Plugins can be installed as either single `.lua` files or `.zip` archives. Lector automatically organizes all plugins into dedicated subdirectories (e.g., `plugins/my-plugin/init.lua`). For more details, see [API.md](API.md).
