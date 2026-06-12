# Lector Plugin API Documentation

This document is the single source of truth for the Lector Plugin API. All Lua-exposed APIs are documented here.

## Security & Capabilities

Plugins must enable capabilities using `app.register_manifest(manifest)` or `app.enable_capability(name)` before they can access restricted namespaces.

### Available Capabilities

| Capability         | Description                                                    |
| :----------------- | :------------------------------------------------------------- |
| `ui`               | Access to UI extension APIs (`add_tab`, `add_section`, etc.)   |
| `theming`          | Access to `app.ui.add_style` for custom CSS                    |
| `network`          | Access to network functions and domain permissions             |
| `local_network`    | Allows network requests to local/private IPs (if enabled)      |
| `storage`          | Access to the persistent key-value store and jailed filesystem |
| `source`           | Required for document provider plugins (Search/Get Document)   |
| `catalog`          | Required to appear in the global Discovery/Search interface    |
| `background`       | Access to background tasks (`app.spawn`, `app.sleep`)          |
| `interaction`      | Access to `app.rpc` for inter-plugin communication             |
| `global_documents` | Access to documents across the entire library (not just owned) |

---

## Sandbox Restrictions

Lector executes plugins in a hardened Lua sandbox.

- **Whitelisted Modules**: `math`, `string`, `table`, `os`.
- **Restricted `os` Module**: Only `os.time`, `os.date`, `os.difftime`, and `os.clock` are available. `os.execute`, `os.rename`, etc., are **DISABLED**.
- **Disabled Modules**: `io`, `package`, `debug`, `require`, `module`.
- **Network Isolation**: All requests are checked against an IP-level SSRF filter. Access to local/private networks is **BLOCKED** unless the `local_network` capability is explicitly granted.

---

## Global Functions & Extensions

### `print(message: string)`

Logs a timestamped message to the server terminal.

### `url_encode(str: string) -> string`

Returns a URL-encoded version of the string.

### `url_join(base: string, relative: string) -> string`

Safely joins a base URL and a relative path.

### `json_decode(json_string: string) -> table | nil`

Parses a JSON string into a Lua table. Returns `nil` on failure.

### `json_encode(data: table) -> string`

Converts a Lua table into a JSON string.

### `css_select(html: string, selector: string) -> table[]`

Parses HTML and returns an array of matching elements. Each element contains `html`, `text`, `href`, and `attrs` (a table of HTML attributes).

### String Extensions

Lector extends the default Lua string prototype with useful helpers:

- `str:contains(substr: string) -> boolean`: Checks if the string contains a substring.
- `str:trim() -> string`: Removes leading and trailing whitespace.
- `str:split(separator: string) -> table`: Splits the string into an array by the separator.

---

## `app` Namespace

### `app.set_id(id: string)`

Defines the plugin's unique identifier (lowercase, no spaces).

### `app.register_manifest(manifest: table)`

Registers the plugin's architectural type and capabilities. This is the preferred way to initialize a plugin.
**Required Fields**:

- `type` (string): The type of the plugin (`source` or `utility`). Default is `source`.
- `capabilities` (table): An array of strings defining required capabilities.

Example:

```lua
app.register_manifest({
    type = "utility",
    capabilities = {"ui", "network"}
})
```

### `app.enable_capability(name: string)`

Enables a single security capability. (Legacy, prefer `register_manifest`).

### `app.add_permission(domain: string)`

Grants access to a network domain (use `*` for all). Requires `network`.

### `app.add_section(id: string, label: string)`

Adds a sidebar section grouping. Requires `ui`.

### `app.add_tab(id: string, label: string, icon: string, section_id: string, component: string)`

Adds a sidebar tab under a specific section. Requires `ui`.

### `app.add_settings_group(id: string, label: string)`

Registers a settings group that will appear in the global Settings page. Requires `ui`. The plugin must export a `get_settings_schema` function to define the fields.

### `app.add_action(context: string, label: string, method: string, [icon: string])`

Adds an action button to the UI. Contexts: `selection`, `document`, `global`. Requires `ui`.

### `app.rpc(target_plugin: string, method: string, [args_json: string]) -> string`

Calls a function exported by another plugin. Requires `interaction`.

### `app.log(message: string, [level: string])`

Logs a message (INFO, WARN, ERROR) to the server terminal.

### `app.spawn(func_name: string, [args_json: string])`

Starts a background thread in an isolated VM. Requires `background`.

### `app.sleep(ms: number)`

Pauses execution. Use only in spawned threads. Requires `background`.

### `app.store.set(key: string, value: string)`

Saves a string permanently in the database for this plugin. Requires `storage`.

### `app.store.get(key: string) -> string | nil`

Retrieves a stored string. Requires `storage`.

---

## `app.ui` Namespace

### `app.ui.add_style(css: string)`

Injects custom CSS into the frontend. Requires `theming`.

### `app.ui.set_override(key: string, config: table)`

Overrides an internal UI component. Currently supports overriding the `reader` via custom `iframe` injection. Requires `ui`.

```lua
app.ui.set_override("reader", {
    type = "iframe",
    url = "/api/plugins/myplugin/assets/player.html"
})
```

_Note: Plugins can serve static files by placing them in an `assets/` folder. They are accessible at `/api/plugins/{plugin_id}/assets/{filename}`._

### `app.ui.open_stream(url: string, [headers: table])`

Signals the frontend to open a generic media stream player. Requires `ui`.

### `app.ui.open_gallery(images: string[], [headers: table])`

Signals the frontend to open an image gallery. Requires `ui`.

---

## `doc` Namespace

All `doc` functions are scoped to the plugin's own documents unless the `global_documents` capability is enabled.

### `doc.list() -> table[]`

Returns a list of all documents. Each table element contains `id`, `title`, `url`, `cover_url`, `author`, `studio`, `synopsis`, `genres`, `status`, `type`, and `is_in_library`.

### `doc.get_chapters(doc_id: number) -> table[]`

Returns a list of chapters (`id`, `title`, `url`, `status`, `metadata`).

### `doc.update_chapter_content(chapter_id: number, content: string)`

Saves chapter text content to the database.

### `doc.update_chapter_metadata(chapter_id: number, metadata: string)`

Saves chapter multimedia metadata (e.g., JSON string of image/stream URLs).

### `doc.batch_upsert_chapters(doc_id: number, chapters: table)`

Performs a high-performance bulk UPSERT of chapters for a document. The `chapters` table should be an array of tables, each containing `title`, `url`, `order`, `status`, and `metadata`.

### `doc.update_metadata(doc_id: number, metadata: table)`

Updates document metadata. Supported fields: `title`, `author`, `synopsis`, `genres`, `status`, `cover_url`, `studio`, `type`.

### `doc.clean(html: string, [title: string]) -> string`

Strips ads and junk from HTML using Lector's high-performance cleaner.

### `doc.get_progress(doc_id: number) -> table | nil`

Retrieves the user's reading/viewing progress for a document.

### `doc.set_progress(doc_id: number, progress: table)`

Sets the user's reading progress. The `progress` table expects `chapter_id`, `scroll_pos`, and `client_updated_at`.

### `doc.fetch_chapter(source: string, url: string) -> table | nil`

Triggers the source plugin to fetch chapter data (content or metadata) for a given URL.

### `doc.export_to(doc_id: number, format: string, dest_path: string) -> boolean`

Exports a document to a specific format (e.g., "epub") and saves it to `dest_path`. Requires `storage`.

### `doc.write_to(dest_path: string, content: string) -> boolean`

Writes raw content to a file in the downloads directory. Requires `storage`.

---

## `net` Namespace

Requires `network` capability. Most functions are accessible via `net.` or `app.net.`.

### `net.fetch_retry(method: string, url: string, [options: table]) -> string`

Performs a network request with built-in exponential backoff (retries up to 3 times, with increasing delays). Options: `body`. Returns the body or an "ERROR:" prefix. Requires `network`.

### `net.request(method: string, url: string, [options: table]) -> string`

Performs a network request with advanced options. (Also available as `app.net.request`).
Options:

- `body`: POST data string.
- `referer`: Custom referer.
- `headers`: Table of custom headers.
- `is_ajax`: Set `X-Requested-With` header.

### `net.fetch(url: string) -> string`

Performs a simple GET request and returns the body.

### `net.download(url: string, dest_path: string, [options: table]) -> boolean`

Downloads a file directly to the local filesystem. Options supports `headers` and `referer`. Requires `storage`.

### `net.set_profile(profile: string)`

Selects a generic networking profile (e.g., `standard`, `mobile`) to ensure compatibility.

### `net.url_encode(str: string) -> string`

Returns a URL-encoded version of the string.

### `net.url_decode(str: string) -> string`

Returns a URL-decoded version of the string.

## `fs` Namespace

Restricted to `/app/data/plugins/<plugin_name>/`. Requires `storage`.

### `fs.read_file(path: string) -> string | nil`

Reads a file from the jailed directory.

### `fs.write_file(path: string, content: string) -> boolean`

Writes a file to the jailed directory.
