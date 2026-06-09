app.set_id("scroll")
app.enable_capability("ui")

app.ui.set_override("reader", {
    type = "iframe",
    url = "/api/plugins/scroll/assets/player.html"
})
