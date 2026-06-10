app.set_id("hello_world")
app.register_manifest({type="utility"})
app.enable_capability("ui")

app.add_section("examples", "Examples")
app.add_tab("hello_world", "Hello World", "Zap", "examples", "dynamic")

function get_ui_schema(args_json)
    return {
        title = "Hello Lector",
        subtitle = "A simple example plugin demonstrating the UI API",
        components = {
            {
                id = "welcome_text",
                type = "Text",
                props = {
                    title = "Welcome to the Plugin API",
                    text = "This page is rendered entirely from a Lua script. You can use components like SortableList, Text, and Header to build rich interfaces."
                }
            }
        }
    }
end
