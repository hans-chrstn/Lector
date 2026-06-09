package plugin

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/lector/internal/binder"
	"github.com/user/lector/internal/core/sanitizer"
	lua "github.com/yuin/gopher-lua"
	"os"
	"path/filepath"
)

func (s *LuaPlugin) registerDocFunctions() {
	s.L.SetGlobal("css_select", s.L.NewFunction(s.cssSelect))

	doc := s.L.NewTable()
	s.L.SetField(doc, "clean", s.L.NewFunction(s.docClean))
	s.L.SetField(doc, "update_chapter_content", s.L.NewFunction(s.docUpdateChapterContent))
	s.L.SetField(doc, "update_chapter_metadata", s.L.NewFunction(s.docUpdateChapterMetadata))
	s.L.SetField(doc, "update_metadata", s.L.NewFunction(s.docUpdateMetadata))
	s.L.SetField(doc, "get_chapters", s.L.NewFunction(s.docGetChapters))
	s.L.SetField(doc, "list", s.L.NewFunction(s.docList))
	s.L.SetField(doc, "get_progress", s.L.NewFunction(s.docGetProgress))
	s.L.SetField(doc, "set_progress", s.L.NewFunction(s.docSetProgress))
	s.L.SetField(doc, "export_to", s.L.NewFunction(s.docExportTo))
	s.L.SetField(doc, "fetch_chapter", s.L.NewFunction(s.docFetchChapter))
	s.L.SetField(doc, "write_to", s.L.NewFunction(s.docWriteTo))
	s.L.SetGlobal("doc", doc)
}

func (s *LuaPlugin) docUpdateMetadata(L *lua.LState) int {
	docID := L.CheckInt(1)
	meta := L.CheckTable(2)

	updates := make(map[string]interface{})
	meta.ForEach(func(k, v lua.LValue) {
		key := k.String()
		val := ""
		if v.Type() != lua.LTNil {
			val = v.String()
		}
		if key == "type" || key == "title" || key == "author" || key == "synopsis" || key == "genres" || key == "status" || key == "cover_url" || key == "studio" {
			updates[key] = val
		}
	})

	success := s.Store.UpdateDocumentMetadata(docID, s.Name, s.HasCapability("global_documents"), updates)
	L.Push(lua.LBool(success))
	return 1
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

func (s *LuaPlugin) docUpdateChapterMetadata(L *lua.LState) int {
	chapterID := L.CheckInt(1)
	metadata := L.CheckString(2)

	success := s.Store.UpdateChapterMetadata(chapterID, s.Name, s.HasCapability("global_documents"), metadata)
	L.Push(lua.LBool(success))
	return 1
}

func (s *LuaPlugin) docUpdateChapterContent(L *lua.LState) int {
	chapterID := L.CheckInt(1)
	content := L.CheckString(2)

	success := s.Store.UpdateChapterContent(chapterID, s.Name, s.HasCapability("global_documents"), content)
	L.Push(lua.LBool(success))
	return 1
}

func (s *LuaPlugin) docGetChapters(L *lua.LState) int {
	docID := L.CheckInt(1)

	chapters, ok := s.Store.GetChapters(docID, s.Name, s.HasCapability("global_documents"))
	if !ok {
		L.Push(L.NewTable())
		return 1
	}

	res := L.NewTable()
	for _, ch := range chapters {
		it := L.NewTable()
		it.RawSetString("id", lua.LNumber(ch.ID))
		it.RawSetString("title", lua.LString(ch.Title))
		it.RawSetString("url", lua.LString(ch.URL))
		it.RawSetString("status", lua.LString(ch.Status))
		it.RawSetString("metadata", lua.LString(ch.Metadata))
		res.Append(it)
	}
	L.Push(res)
	return 1
}

func (s *LuaPlugin) docClean(L *lua.LState) int {
	html := L.CheckString(1)
	title := L.OptString(2, "")
	L.Push(lua.LString(sanitizer.CleanHTML(html, title)))
	return 1
}

func (s *LuaPlugin) docList(L *lua.LState) int {
	documents := s.Store.ListDocuments(s.Name, s.HasCapability("global_documents"))

	res := L.NewTable()
	for _, doc := range documents {
		it := L.NewTable()
		it.RawSetString("id", lua.LNumber(doc.ID))
		it.RawSetString("title", lua.LString(doc.Title))
		it.RawSetString("url", lua.LString(doc.URL))
		it.RawSetString("cover_url", lua.LString(doc.CoverURL))
		it.RawSetString("author", lua.LString(doc.Author))
		it.RawSetString("studio", lua.LString(doc.Studio))
		it.RawSetString("synopsis", lua.LString(doc.Synopsis))
		it.RawSetString("genres", lua.LString(doc.Genres))
		it.RawSetString("status", lua.LString(doc.Status))
		it.RawSetString("type", lua.LString(doc.Type))
		it.RawSetString("is_in_library", lua.LBool(doc.IsInLibrary))
		res.Append(it)
	}
	L.Push(res)
	return 1
}

func (s *LuaPlugin) docGetProgress(L *lua.LState) int {
	docID := L.CheckInt(1)
	prog, ok := s.Store.GetReadingProgress(docID)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}

	res := L.NewTable()
	res.RawSetString("chapter_id", lua.LNumber(prog.ChapterID))
	res.RawSetString("scroll_pos", lua.LNumber(prog.ScrollPos))
	res.RawSetString("client_updated_at", lua.LNumber(prog.ClientUpdatedAt))
	L.Push(res)
	return 1
}

func (s *LuaPlugin) docSetProgress(L *lua.LState) int {
	docID := L.CheckInt(1)
	progTable := L.CheckTable(2)

	var chapterID *uint
	if val := progTable.RawGetString("chapter_id"); val != lua.LNil {
		if n, ok := val.(lua.LNumber); ok {
			ch := uint(n)
			chapterID = &ch
		}
	}
	var scrollPos *float64
	if val := progTable.RawGetString("scroll_pos"); val != lua.LNil {
		if n, ok := val.(lua.LNumber); ok {
			sp := float64(n)
			scrollPos = &sp
		}
	}
	var clientUpdatedAt *int64
	if val := progTable.RawGetString("client_updated_at"); val != lua.LNil {
		if n, ok := val.(lua.LNumber); ok {
			cu := int64(n)
			clientUpdatedAt = &cu
		}
	}

	success := s.Store.SetReadingProgress(docID, chapterID, scrollPos, clientUpdatedAt)
	L.Push(lua.LBool(success))
	return 1
}

func (s *LuaPlugin) docExportTo(L *lua.LState) int {
	if !s.HasCapability("storage") {
		L.Push(lua.LBool(false))
		return 1
	}

	docID := L.CheckInt(1)
	format := L.CheckString(2)
	destPath := L.CheckString(3)

	doc, ok := s.Store.GetDocumentForExport(docID)
	if !ok {
		L.Push(lua.LBool(false))
		return 1
	}

	baseDir, _ := filepath.Abs("downloads")
	baseDir = filepath.Clean(baseDir)
	fullPath, _ := filepath.Abs(destPath)
	fullPath = filepath.Clean(fullPath)

	if fullPath != baseDir && !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) {
		L.Push(lua.LBool(false))
		return 1
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		L.Push(lua.LBool(false))
		return 1
	}

	var binderErr error
	if format == "pdf" {
		binderErr = binder.BindPDF(doc, fullPath)
	} else {
		binderErr = binder.BindEPUB(doc, fullPath)
	}

	L.Push(lua.LBool(binderErr == nil))
	return 1
}

func (s *LuaPlugin) docFetchChapter(L *lua.LState) int {
	sourceName := L.CheckString(1)
	chapterURL := L.CheckString(2)

	PluginsMu.Lock()
	p, ok := GlobalPlugins[sourceName]
	PluginsMu.Unlock()

	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("Plugin not found"))
		return 2
	}

	ch, err := p.GetChapter(chapterURL)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	res := L.NewTable()
	res.RawSetString("title", lua.LString(ch.Title))
	res.RawSetString("content", lua.LString(ch.Content))
	res.RawSetString("metadata", lua.LString(ch.Metadata))
	L.Push(res)
	return 1
}

func (s *LuaPlugin) docWriteTo(L *lua.LState) int {
	if !s.HasCapability("storage") {
		L.Push(lua.LBool(false))
		return 1
	}
	destPath := L.CheckString(1)
	content := L.CheckString(2)
	baseDir, _ := filepath.Abs("downloads")
	baseDir = filepath.Clean(baseDir)
	fullPath, _ := filepath.Abs(destPath)
	fullPath = filepath.Clean(fullPath)
	if fullPath != baseDir && !strings.HasPrefix(fullPath, baseDir+string(filepath.Separator)) {
		L.Push(lua.LBool(false))
		return 1
	}
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		L.Push(lua.LBool(false))
		return 1
	}
	err := os.WriteFile(fullPath, []byte(content), 0644)
	L.Push(lua.LBool(err == nil))
	return 1
}
