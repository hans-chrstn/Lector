package plugin

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) registerDocFunctions() {
	s.L.SetGlobal("css_select", s.L.NewFunction(s.cssSelect))

	doc := s.L.NewTable()
	s.L.SetField(doc, "clean", s.L.NewFunction(s.docClean))
	s.L.SetField(doc, "update_chapter_content", s.L.NewFunction(s.docUpdateChapterContent))
	s.L.SetField(doc, "get_chapters", s.L.NewFunction(s.docGetChapters))
	s.L.SetGlobal("doc", doc)
}

func (s *LuaPlugin) cssSelect(L *lua.LState) int {
	html, selector := L.CheckString(1), L.CheckString(2)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	res := L.NewTable()
	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		it := L.NewTable()
		it.RawSetString("text", lua.LString(strings.TrimSpace(sel.Text())))
		h, _ := sel.Html()
		it.RawSetString("html", lua.LString(h))
		href, _ := sel.Attr("href")
		it.RawSetString("href", lua.LString(href))
		attrs := L.NewTable()
		if len(sel.Nodes) > 0 {
			for _, a := range sel.Nodes[0].Attr {
				attrs.RawSetString(a.Key, lua.LString(a.Val))
			}
		}
		it.RawSetString("attrs", attrs)
		res.Append(it)
	})
	L.Push(res)
	return 1
}

func (s *LuaPlugin) docUpdateChapterContent(L *lua.LState) int {
	chapterID := L.CheckInt(1)
	content := L.CheckString(2)

	result := db.DB.Model(&models.Chapter{}).Where("id = ?", chapterID).Updates(map[string]interface{}{
		"content": content,
		"status":  "done",
	})

	L.Push(lua.LBool(result.Error == nil))
	return 1
}

func (s *LuaPlugin) docGetChapters(L *lua.LState) int {
	docID := L.CheckInt(1)
	var chapters []models.Chapter
	db.DB.Where("document_id = ?", docID).Order("`order` ASC").Find(&chapters)

	res := L.NewTable()
	for _, ch := range chapters {
		it := L.NewTable()
		it.RawSetString("id", lua.LNumber(ch.ID))
		it.RawSetString("title", lua.LString(ch.Title))
		it.RawSetString("url", lua.LString(ch.URL))
		it.RawSetString("status", lua.LString(ch.Status))
		res.Append(it)
	}
	L.Push(res)
	return 1
}

func (s *LuaPlugin) docClean(L *lua.LState) int {
	html := L.CheckString(1)
	title := L.OptString(2, "")
	L.Push(lua.LString(CleanHTML(html, title)))
	return 1
}
