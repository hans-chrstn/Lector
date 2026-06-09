app.set_id("translate")
app.enable_capability("ui")
app.enable_capability("network")
app.enable_capability("storage")
app.add_permission("*")

app.add_action("selection", "Translate Selection", "translate_text", "Globe")

app.add_settings_group("translate_settings", "Translate Configuration")

exports = {
	get_settings_schema = function(group_id_json)
		local group_id = json_decode(group_id_json)
		if group_id == "translate_settings" then
			return json_encode({
				title = "Translation Settings",
				subtitle = "Configure the source and target languages for translation.",
				components = {
					{
						id = "source_lang",
						type = "TextInput",
						props = {
							label = "Source Language (e.g. auto, ja, en, es)",
							defaultValue = app.store.get("source_lang") or "auto",
						},
					},
					{
						id = "target_lang",
						type = "TextInput",
						props = {
							label = "Target Language (e.g. en, es, fr, de)",
							defaultValue = app.store.get("target_lang") or "en",
						},
					},
					{
						id = "save_btn",
						type = "Button",
						props = { label = "Save Settings", method = "save_settings" },
					},
				},
			})
		end
		return "{}"
	end,

	save_settings = function(args_json)
		local args = json_decode(args_json)
		app.store.set("source_lang", args.source_lang or "auto")
		app.store.set("target_lang", args.target_lang or "en")
		return '{"status":"success","message":"Translation settings saved!"}'
	end,

	translate_text = function(args_json)
		local data = json_decode(args_json)
		if not data then
			return '{"error":"Invalid context"}'
		end

		local text = ""
		if type(data) == "string" then
			text = data
		elseif type(data) == "table" then
			text = data.synopsis or ""
		end

		if text == "" then
			return '{"error":"No text provided to translate."}'
		end

		local source_lang = app.store.get("source_lang") or "auto"
		local target_lang = app.store.get("target_lang") or "en"

		local lang_names = {
			en = "English",
			es = "Spanish",
			ja = "Japanese",
			fr = "French",
			de = "German",
			zh = "Chinese",
			ru = "Russian",
			pt = "Portuguese",
			it = "Italian",
			ko = "Korean",
			auto = "Auto",
		}

		local encoded_text = url_encode(text)
		local url = "https://translate.googleapis.com/translate_a/single?client=gtx&sl="
			.. source_lang
			.. "&tl="
			.. target_lang
			.. "&dt=t&q="
			.. encoded_text

		local res = app.net.request("GET", url)
		if not res then
			return '{"error":"Failed to connect to Google Translate service."}'
		end

		local parsed = json_decode(res)
		if parsed and parsed[1] then
			local translated = ""
			for _, sentence in ipairs(parsed[1]) do
				if sentence[1] then
					translated = translated .. sentence[1]
				end
			end

			local detected_lang = source_lang
			if source_lang == "auto" then
				for i = 1, 5 do
					if type(parsed[i]) == "string" then
						detected_lang = parsed[i]
						break
					end
				end
			end

			local source_name = lang_names[detected_lang] or detected_lang
			local target_name = lang_names[target_lang] or target_lang

			local result_obj = {
				title = "Translation",
				phonetic = source_name .. " ➔ " .. target_name,
				meanings = {
					{
						partOfSpeech = "Result",
						definitions = {
							{ definition = translated },
						},
					},
				},
			}
			return json_encode(result_obj)
		else
			return '{"error":"Failed to parse Google Translate API response."}'
		end
	end,
}
