# Lector Plugin API Documentation

This document is the single source of truth for the Lector Plugin API. All Lua-exposed APIs are documented here.

## Security & Capabilities

Plugins must enable capabilities using `app.enable_capability(name)` before they can access restricted namespaces.

### Available Capabilities

| Capability    | Description                                                    |
| :------------ | :------------------------------------------------------------- |
| `ui`          | Access to UI extension APIs (`add_tab`, `add_section`, etc.)   |
| `theming`     | Access to `app.ui.add_style` for custom CSS                    |
| `network`     | Access to network functions and domain permissions             |
| `storage`     | Access to the persistent key-value store and jailed filesystem |
| `source`      | Required for document provider plugins (Search/Get Document)   |
| `catalog`     | Required to appear in the global Discovery/Search interface    |
| `background`  | Access to background tasks (`app.spawn`, `app.sleep`)          |
| `interaction` | Access to `app.rpc` for inter-plugin communication             |

---

## Sandbox Restrictions

Lector executes plugins in a hardened Lua sandbox.

- **Whitelisted Modules**: `math`, `string`, `table`, `os`.
- **Restricted `os` Module**: Only `os.time`, `os.date`, `os.difftime`, and `os.clock` are available. `os.execute`, `os.rename`, etc., are **DISABLED**.
- **Disabled Modules**: `io`, `package`, `debug`, `require`, `module`.
- **Network Isolation**: All requests are checked against an IP-level SSRF filter. Access to local/private networks (localhost, 192.168.x.x, etc.) is **BLOCKED**.

---

## Global Functions

### `print(message: string)`
Logs a timestamped message to the server terminal.

### `url_encode(str: string) -> string`
Returns a URL-encoded version of the string.

### `url_join(base: string, relative: string) -> string`
Safely joins a base URL and a relative path.

---

## `app` Namespace

### `app.set_id(id: string)`
Defines the plugin's unique identifier (lowercase, no spaces).

### `app.enable_capability(name: string)`
Enables a security capability.

### `app.add_permission(domain: string)`
Grants access to a network domain (use `*` for all). Requires `network`.

### `app.ui.set_override(key: string, config: table)`
Overrides an internal UI component.
Currently supports overriding the `reader` via custom `iframe` injection.
```lua
app.ui.set_override("reader", {
    type = "iframe",
    url = "/api/plugins/myplugin/assets/player.html"
})
```
*Note: Plugins can serve static files by placing them in an `assets/` folder alongside their `init.lua`. They are accessible at `/api/plugins/{plugin_id}/assets/{filename}`.*

### `app.add_section(id: string, label: string)`
Adds a sidebar section. Requires `ui`.

### `app.add_tab(id: string, label: string, icon: string, section_id: string, component: string)`
Adds a sidebar tab. Requires `ui`.

### `app.add_action(context: string, label: string, method: string, [icon: string])`
Adds an action button to the UI. Contexts: `selection`, `document`, `global`. Requires `ui`.

### `app.rpc(target_plugin: string, method: string, [args_json: string]) -> string`
Calls a function in another plugin. Requires `interaction`.

### `app.log(message: string, [level: string])`
Logs a message (INFO, WARN, ERROR).

### `app.spawn(func_name: string, [args_json: string])`
Starts a background thread in an isolated VM. Requires `background`.

### `app.sleep(ms: number)`
Pauses execution. Use only in spawned threads. Requires `background`.

### `app.store.set(key: string, value: string)`
Saves a string permanently in the database. Requires `storage`.

### `app.store.get(key: string) -> string | nil`
Retrieves a stored string. Requires `storage`.

### `app.net.request(method: string, url: string, [options: table]) -> string`
Performs a network request. Options: `body`, `referer`, `is_ajax`. Requires `network`.

---

## `app.ui` Namespace

### `app.ui.add_style(css: string)`
Injects custom CSS into the frontend. Requires `theming`.

### `app.ui.set_override(key: string, values: table)`
Overrides internal UI configurations. Requires `ui`.

### `app.ui.open_stream(url: string, [headers: table])`
Signals the frontend to open a generic media stream player. Requires `ui`.

### `app.ui.open_gallery(images: string[], [headers: table])`
Signals the frontend to open an image gallery. Requires `ui`.

---

## `doc` Namespace

All `doc` functions are scoped to the plugin's own documents.

### `doc.list() -> table[]`
Returns a list of all documents owned by the calling plugin. Each table element contains `id`, `title`, `url`, `cover_url`, `author`, `studio`, `synopsis`, `genres`, `status`, `type`, and `is_in_library`.

### `doc.get_chapters(doc_id: number) -> table[]`
Returns list of chapters (`id`, `title`, `url`, `status`, `metadata`).

### `doc.update_chapter_content(chapter_id: number, content: string)`
Saves chapter text content to the database. Only works for documents owned by the plugin.

### `doc.update_chapter_metadata(chapter_id: number, metadata: string)`
Saves chapter multimedia metadata (e.g. JSON string of image/stream URLs). Only works for documents owned by the plugin.

### `doc.update_metadata(doc_id: number, metadata: table)`
Updates synopsis, genres, status, cover, or type. Supported fields: `title`, `author`, `synopsis`, `genres`, `status`, `cover_url`, `studio`, `type`.
Valid `type` values: `text`, `images`, `stream`.


### `doc.clean(html: string, [title: string]) -> string`
Strips ads and junk from HTML using Lector's high-performance cleaner.

### `doc.write_to(dest_path: string, content: string) -> boolean`
Writes raw content to a file in the downloads directory. Requires `storage`.

---

## `net` Namespace

Requires `network` capability.

### `app.net.fetch(url: string) -> string`
Performs a GET request and returns the body.

### `app.net.request(method: string, url: string, [options: table]) -> string`
Performs a network request with advanced options. 
Options:
- `body`: POST data.
- `referer`: Custom referer.
- `is_ajax`: Set `X-Requested-With` header.

### `app.net.set_profile(profile: string)`
Selects a generic networking profile to ensure compatibility with diverse web infrastructure.
Valid profiles: `standard` (Desktop), `mobile`.

---

## Request Orchestration

Lector automatically manages request cadence to ensure natural intervals and prevent server strain on remote providers. This includes randomized delays (jitter) between sequential requests from the same plugin.

---

## `fs` Namespace (Jailed)

Restricted to `/app/data/plugins/<plugin_name>/`. Requires `storage`.

### `fs.read_file(path: string) -> string | nil`
Reads a file from the jailed directory.

### `fs.write_file(path: string, content: string) -> boolean`
Writes a file to the jailed directory.

