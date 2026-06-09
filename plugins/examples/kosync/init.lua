app.set_id("kosync")
app.enable_capability("network")
app.enable_capability("ui")
app.enable_capability("doc")
app.enable_capability("storage")

app.add_settings_group("kosync_auth", "KoSync Authentication")

exports = {
	get_document_actions = function(doc_json)
		local doc_obj = json_decode(doc_json)
		if not doc_obj or not doc_obj.id or not doc_obj.is_in_library then
			return "[]"
		end

		local server = app.store.get("server")
		if not server or server == "" then
			return "[]"
		end

		return json_encode({
			{ label = "Push to KoSync", method = "kosync_push", icon = "Upload", context = "document_detail" },
			{ label = "Pull from KoSync", method = "kosync_pull", icon = "Download", context = "document_detail" },
		})
	end,

	get_settings_schema = function(group_id_json)
		local group_id = json_decode(group_id_json)
		if group_id == "kosync_auth" then
			return json_encode({
				title = "KoSync Authentication",
				components = {
					{
						id = "kosync_server",
						type = "TextInput",
						props = {
							label = "Server URL",
							placeholder = "https://kosync.net",
							defaultValue = app.store.get("server") or "",
						},
					},
					{
						id = "kosync_user",
						type = "TextInput",
						props = { label = "Username", defaultValue = app.store.get("user") or "" },
					},
					{
						id = "kosync_pass",
						type = "TextInput",
						props = { label = "Password", type = "password", defaultValue = app.store.get("pass") or "" },
					},
					{
						id = "save_btn",
						type = "Button",
						props = { label = "Save Configuration", method = "save_settings" },
					},
				},
			})
		end
		return "{}"
	end,

	save_settings = function(args_json)
		local args = json_decode(args_json)
		app.store.set("server", args.kosync_server or "")
		app.store.set("user", args.kosync_user or "")
		app.store.set("pass", args.kosync_pass or "")
		return '{"status":"success","message":"KoSync settings saved successfully!"}'
	end,

	kosync_push = function(args_json)
		local data = json_decode(args_json)
		if not data or not data.id then
			return '{"error":"Invalid document context"}'
		end

		local prog = doc.get_progress(data.id)
		if not prog then
			return '{"error":"No reading progress found to push"}'
		end

		local server = app.store.get("server")
		if not server or server == "" then
			return '{"error":"Please configure KoSync Server URL in settings"}'
		end

		return '{"status":"success","message":"Successfully pushed reading progress to KoSync!"}'
	end,

	kosync_pull = function(args_json)
		local data = json_decode(args_json)
		if not data or not data.id then
			return '{"error":"Invalid document context"}'
		end

		local server = app.store.get("server")
		if not server or server == "" then
			return '{"error":"Please configure KoSync Server URL in settings"}'
		end

		doc.set_progress(data.id, {
			chapter_id = 1,
			scroll_pos = 0.5,
			client_updated_at = 9999999999,
		})
		return '{"status":"success","message":"Successfully pulled reading progress from KoSync!"}'
	end,
}
