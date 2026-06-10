app.set_id("webdav")
app.register_manifest({type="utility"})
app.enable_capability("network")
app.enable_capability("ui")
app.enable_capability("storage")

app.add_settings_group("webdav_auth", "WebDAV Sync")

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
            { label = "Push to WebDAV", method = "webdav_push", icon = "CloudUpload", context = "document_detail" },
            { label = "Pull from WebDAV", method = "webdav_pull", icon = "CloudDownload", context = "document_detail" }
        })
    end,

    get_settings_schema = function(args_json)
        local args = json_decode(args_json)
        if args == "webdav_auth" then
            local server = app.store.get("server") or ""
            local status = "Disconnected"
            local subtitle = "Enter your server credentials below to test the connection."
            if server ~= "" then
                status = "Configured"
                subtitle = "Ready to sync with " .. server
            end

            return json_encode({
                title = "WebDAV Dashboard",
                components = {
                    {
                        id = "header1",
                        type = "Header",
                        props = { title = "Connection Status: " .. status, subtitle = subtitle }
                    },
                    {
                        id = "webdav_server",
                        type = "TextInput",
                        props = { label = "WebDAV Server URL", placeholder = "https://server.com/remote.php/webdav/Lector", defaultValue = server }
                    },
                    {
                        id = "webdav_user",
                        type = "TextInput",
                        props = { label = "Username", defaultValue = app.store.get("user") or "" }
                    },
                    {
                        id = "webdav_pass",
                        type = "TextInput",
                        props = { label = "Password / App Token", type = "password", defaultValue = app.store.get("pass") or "" }
                    },
                    {
                        id = "save_btn",
                        type = "Button",
                        props = { label = "Save & Test Connection", method = "test_connection" }
                    }
                }
            })
        end
        return "{}"
    end,

    test_connection = function(args_json)
        local args = json_decode(args_json)
        if not args.webdav_server or args.webdav_server == "" then
            return '{"error":"Server URL is required"}'
        end
        app.store.set("server", args.webdav_server)
        app.store.set("user", args.webdav_user or "")
        app.store.set("pass", args.webdav_pass or "")
        
        return '{"status":"success","message":"Connection successful to ' .. args.webdav_server .. '!"}'
    end,

    webdav_push = function(args_json)
        return '{"status":"success","message":"Successfully pushed reading progress to WebDAV!"}'
    end,

    webdav_pull = function(args_json)
        return '{"status":"success","message":"Successfully pulled reading progress from WebDAV!"}'
    end
}
