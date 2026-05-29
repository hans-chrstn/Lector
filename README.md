# Lector

Lector is a neutral, open-source document management and reading application. It is designed to allow users to organize and read their own locally-stored files (such as EPUB, PDF, and CBZ/CBR) and to extend the application's functionality through a Lua-based plugin system.

---

## Disclaimer and Terms of Use

### 1. Neutral Tool Policy
Lector is a neutral document management tool. It does not provide, host, or distribute any content. The developers of Lector do not provide any site-specific scraping plugins or tools intended to bypass paywalls or access copyrighted material without authorization. Users are solely responsible for the legality of the content they import into Lector and the plugins they choose to install and execute.

### 2. No Endorsement
The inclusion of a plugin system is intended for legitimate customization and interoperability. The developers of Lector do not endorse, support, or encourage the use of Lector for any activities that violate copyright laws or the terms of service of any third-party websites.

### 3. Limitation of Liability
The software is provided "as is", without warranty of any kind. In no event shall the authors or copyright holders be liable for any claim, damages, or other liability arising from the use of the software by end-users, including the use of third-party plugins.

### 4. DMCA Compliance
Lector complies with the Digital Millennium Copyright Act (DMCA). If you believe that your copyrighted work has been infringed by the software itself (the core reader engine), please contact the maintainers. Note that Lector maintainers cannot control or remove content or plugins hosted on third-party systems or installed locally by users.

---

## Plugin API Documentation

Lector provides a powerful Lua-based API for extending the application's capabilities. All plugins run in a secure sandbox and must explicitly request "Capabilities" to access sensitive functions.

### Security and Capabilities

Plugins must enable capabilities using `app.enable_capability(name)` before they can be used.

| Capability    | Description                                                                |
| :------------ | :------------------------------------------------------------------------- |
| `ui`          | Access to UI extension APIs (`add_tab`, `add_section`, `add_action`, etc.) |
| `theming`     | Access to `app.ui.add_style` for custom CSS injection                      |
| `network`     | Access to `http_get` and domain permission management                      |
| `storage`     | Access to `app.store` for persistent key-value data                        |
| `fs`          | Access to the jailed file system (`fs.*`)                                  |
| `source`      | Required for plugins that act as document sources (Search, Get Document)   |
| `catalog`     | Required to appear in the global discovery/search interface                |
| `background`  | Access to `app.spawn` and `app.sleep` for long-running tasks               |
| `interaction` | Access to `app.rpc` for inter-plugin communication                         |

### Global Functions

#### `print(message: string)`
Logs a message to the Lector server terminal with a timestamp and plugin name.

#### `url_encode(str: string) -> string`
Returns a URL-encoded version of the string.

#### `url_join(base: string, relative: string) -> string`
Safely joins a base URL and a relative path.

#### `css_select(html: string, selector: string) -> table[]`
Parses HTML and returns a list of elements matching the CSS selector. Each element is a table:
- `text`: Inner text of the element.
- `html`: Inner HTML of the element.
- `href`: Value of the `href` attribute (if any).
- `attrs`: Table of all attributes.

#### `http_get(url: string, [referer: string, [is_ajax: boolean]]) -> string`
Fetches a URL and returns the raw HTML/body.
- **Security**: Requires `network` capability and a matching permission via `app.add_permission`.
- **SSRF Protection**: Blocked from accessing private/local subnets.

---

## app Namespace

#### `app.enable_capability(name: string)`
Enables a specific capability for the plugin.

#### `app.add_permission(domain: string)`
Grants the plugin access to a specific network domain. Requires `network` capability.

#### `app.add_section(id: string, label: string)`
Adds a new section to the sidebar. Requires `ui` capability.

#### `app.add_tab(id: string, label: string, icon: string, section_id: string, component: string)`
Adds a new tab to a section. `component` is usually `"dynamic"`.

#### `app.rpc(target_plugin: string, method: string, [args_json: string]) -> string, error`
Calls a function in another plugin. Requires `interaction` capability.

#### `app.log(message: string, [level: string])`
Structured logging. `level` defaults to `"INFO"`.

#### `app.spawn(func_name: string, [args_json: string])`
Starts a background thread. Requires `background` capability.

#### `app.store.set(key: string, value: string)`
Stores a string permanently. Requires `storage` capability.

---

## doc Namespace

#### `doc.get_chapters(doc_id: number) -> table[]`
Returns a list of chapters for a document.

#### `doc.update_chapter_content(chapter_id: number, content: string)`
Updates a chapter's content.

#### `doc.update_metadata(doc_id: number, metadata: table)`
Updates document fields: `synopsis`, `genres`, `status`, `cover_url`.

#### `doc.clean(html: string, [title: string]) -> string`
Runs Lector's high-performance cleaner to strip ads and junk.

---

## fs Namespace (Jailed)

All `fs` functions are restricted to the plugin's private directory: `/app/data/plugins/<plugin_name>/`. Requires `fs` capability.

- `fs.read(path: string) -> string`
- `fs.write(path: string, content: string)`
- `fs.list(path: string) -> table[]`
- `fs.delete(path: string)`
- `fs.exists(path: string) -> boolean`
- `fs.mkdir(path: string)`
