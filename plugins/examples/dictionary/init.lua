app.set_id("dictionary")
local BASE_URL = "https://api.dictionaryapi.dev/api/v2/entries/en/"

app.register_manifest({type="utility"})
app.enable_capability("network")
app.enable_capability("interaction")
app.enable_capability("ui")
app.add_permission("*")

app.add_action("selection", "Define", "define", "Book")

function define(args_json)
    local word = args_json
    if type(args_json) == "string" then
        local data = json_decode(args_json)
        if type(data) == "string" then word = data end
    end
    
    if not word or word == "" then return '{"error": "No word provided"}' end
    
    local clean_word = string.match(word, "([%a%-]+)")
    if not clean_word then return '{"error": "Invalid word"}' end
    
    app.log("Defining word: " .. clean_word)
    local res = app.net.request("GET", BASE_URL .. url_encode(clean_word))
    
    if not res or res == "" or string.find(res, "No Definitions Found") then
        return '{"error": "Definition not found"}'
    end
    
    return res
end

exports = {
    define = define
}
