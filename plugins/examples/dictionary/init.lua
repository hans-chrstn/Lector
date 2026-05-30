app.set_id("dictionary")
local BASE_URL = "https://api.dictionaryapi.dev/api/v2/entries/en/"

app.enable_capability("network")
app.enable_capability("interaction")
app.enable_capability("ui")

app.add_action("selection", "Define", "define", "Book")

function define(word)
    if not word or word == "" then return '{"error": "No word provided"}' end
    
    local clean_word = string.match(word, "([%a%-]+)")
    if not clean_word then return '{"error": "Invalid word"}' end
    
    app.log("Defining word: " .. clean_word)
    local res = http_get(BASE_URL .. url_encode(clean_word))
    
    if not res or res == "" or string.contains(res, "title") and string.contains(res, "No Definitions Found") then
        return '{"error": "Definition not found"}'
    end
    
    return res
end
