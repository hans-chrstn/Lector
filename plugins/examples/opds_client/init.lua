app.set_id("opds_client")
local BASE_URL = "https://standardebooks.org/opds/all"

app.enable_capability("network")
app.enable_capability("ui")
app.enable_capability("catalog")
app.enable_capability("source")

app.add_section("discover", "Discovery")
app.add_tab("opds_catalogs", "Browse catalogs", "Globe", "discover", "dynamic")

function get_ui_schema(args_json)
    local args = {}
    
    local xml = http_get(BASE_URL)
    if not xml or xml == "" then
        return {
            title = "Standard Ebooks",
            components = {
                {
                    id = "error",
                    type = "Text",
                    props = { text = "Failed to load catalog" }
                }
            }
        }
    end

    local entries = css_select(xml, "entry")
    local items = {}
    
    for _, entry in ipairs(entries) do
        local title = css_select(entry.html, "title")[1]
        local author = css_select(entry.html, "author name")[1]
        local cover = css_select(entry.html, "link[rel*='image']")[1]
        local acq = css_select(entry.html, "link[rel*='acquisition']")[1]
        
        table.insert(items, {
            id = acq and acq.attrs["href"] or title.text,
            title = title and title.text or "Unknown",
            subtitle = author and author.text or "Unknown Author",
            cover_url = cover and cover.attrs["href"] or "",
            actions = {
                { label = "Import", method = "import_book", icon = "Download" }
            }
        })
    end

    return {
        title = "Standard Ebooks",
        subtitle = "Browse and import high-quality public domain books",
        components = {
            {
                id = "catalog_list",
                type = "SortableList",
                props = {
                    items = items
                }
            }
        }
    }
end

function import_book(book_url_json)
    app.log("Importing book from: " .. book_url_json)
    return '{"status": "success", "message": "OPDS import acknowledged. Direct download not yet implemented."}'
end
