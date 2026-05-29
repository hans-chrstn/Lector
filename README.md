# Lector

Lector is a neutral, open-source document management and reading application. It is designed to allow users to organize and read their own locally-stored files (such as EPUB, PDF, and CBZ/CBR) and to extend the application's functionality through a Lua-based plugin system.

---

## Security and Terms

Lector is a neutral tool. It does not provide, host, or distribute any content. Users are solely responsible for the legality of the content they import and the plugins they choose to install. For full details, see [DISCLAIMER.md](DISCLAIMER.md).

---

## Deployment

Lector is designed to be easily deployed via Docker. It runs as an unprivileged user for maximum security.

### Docker Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `AUTH_USER` | Username for Basic Authentication | (None) |
| `AUTH_PASSWORD` | Password for Basic Authentication | (None) |
| `MAX_UPLOAD_SIZE` | Maximum upload size in MB | `100` |
| `DATABASE_PATH` | Path to the SQLite database file | `/app/data/lector.db` |
| `CORS_ALLOW_ORIGINS` | Permitted origins for CORS | (Same-origin only) |
| `PORT` | Port to listen on | `3000` |

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

Lector plugins use Lua to extend the application. All plugins run in a secure sandbox and must request capabilities.

### Security and Capabilities

Plugins must enable capabilities using `app.enable_capability(name)` before they can be used.

| Capability | Description |
| :--- | :--- |
| `ui` | Access to UI extension APIs (`add_tab`, `add_section`, etc.) |
| `theming` | Access to `app.ui.add_style` for custom CSS |
| `network` | Access to network functions and domain permissions |
| `storage` | Access to the persistent key-value store and jailed filesystem |
| `source` | Required for document provider plugins (Search/Get Document) |
| `catalog` | Required to appear in the global Discovery/Search interface |
| `background` | Access to background tasks (`app.spawn`, `app.sleep`) |
| `interaction` | Access to `app.rpc` for inter-plugin communication |

### Global Functions

#### `print(message: string)`
Logs a timestamped message to the server terminal.

#### `url_encode(str: string) -> string`
Returns a URL-encoded version of the string.

#### `url_join(base: string, relative: string) -> string`
Safely joins a base URL and a relative path.

#### `css_select(html: string, selector: string) -> table[]`
Parses HTML and returns matching elements. Each element has: `text`, `html`, `href`, and `attrs`.

#### `http_get(url: string, [referer: string, [is_ajax: boolean]]) -> string`
Performs a GET request. Requires `network` capability and domain permission.

#### `http_post(url: string, body: string, [referer: string, [is_ajax: boolean]]) -> string`
Performs a POST request. Requires `network` capability and domain permission.

#### `string.contains(str: string, substr: string) -> boolean`
Returns true if the string contains the substring.

---

### app Namespace

#### `app.enable_capability(name: string)`
Enables a specific security capability.

#### `app.add_permission(domain: string)`
Grants access to a network domain (use `*` for all). Requires `network`.

#### `app.add_section(id: string, label: string)`
Adds a sidebar section. Requires `ui`.

#### `app.add_tab(id: string, label: string, icon: string, section_id: string, component: string)`
Adds a sidebar tab. Requires `ui`.

#### `app.add_action(context: string, label: string, method: string, [icon: string])`
Adds an action button to the UI. Requires `ui`.

#### `app.rpc(target_plugin: string, method: string, [args_json: string]) -> string`
Calls a function in another plugin. Requires `interaction`.

#### `app.log(message: string, [level: string])`
Logs a message (INFO, WARN, ERROR).

#### `app.spawn(func_name: string, [args_json: string])`
Starts a background thread. Requires `background`.

#### `app.sleep(ms: number)`
Pauses execution. Use only in spawned threads. Requires `background`.

#### `app.store.set(key: string, value: string)`
Saves a string permanently in the database. Requires `storage`.

#### `app.store.get(key: string) -> string | nil`
Retrieves a stored string. Requires `storage`.

---

### app.ui Namespace

#### `app.ui.add_style(css: string)`
Injects custom CSS into the frontend. Requires `theming`.

#### `app.ui.set_override(key: string, values: table)`
Overrides internal UI configurations. Requires `ui`.

---

### doc Namespace

#### `doc.get_chapters(doc_id: number) -> table[]`
Returns list of chapters (`id`, `title`, `url`, `status`).

#### `doc.update_chapter_content(chapter_id: number, content: string)`
Saves chapter text to the database.

#### `doc.update_metadata(doc_id: number, metadata: table)`
Updates synopsis, genres, status, or cover.

#### `doc.clean(html: string, [title: string]) -> string`
Strips ads and junk from HTML using Lector's high-performance cleaner.

---

### fs Namespace (Jailed)

All `fs` functions are restricted to `/app/data/plugins/<plugin_name>/`. Requires `storage`.

#### `fs.read_file(path: string) -> string | nil`
Reads a file from the jailed directory.

#### `fs.write_file(path: string, content: string) -> boolean`
Writes a file to the jailed directory.
