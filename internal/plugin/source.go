package plugin

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/lector/internal/models"
	lua "github.com/yuin/gopher-lua"
)

func (s *LuaPlugin) callSearchFunc(name string, param lua.LValue) ([]models.SearchItem, error) {
	pName := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	if !s.HasCapability("source") {
		fmt.Printf("[Security] [%s] Blocked unauthorized source call: %s (Capability 'source' not enabled)\n", pName, name)
		return []models.SearchItem{}, nil
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.L.SetContext(ctx)

	fn := s.L.GetGlobal(name)
	if fn.Type() != lua.LTFunction {
		return []models.SearchItem{}, nil
	}
	if err := s.L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, param); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			pName := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
			fmt.Printf("[Security] [%s] Execution timed out in %s\n", pName, name)
		}
		return []models.SearchItem{}, err
	}
	ret := s.L.Get(-1)
	s.L.Pop(1)
	res := []models.SearchItem{}
	if tbl, ok := ret.(*lua.LTable); ok {
		tbl.ForEach(func(k, v lua.LValue) {
			if it, ok := v.(*lua.LTable); ok {
				res = append(res, models.SearchItem{
					Title:    it.RawGetString("title").String(),
					URL:      it.RawGetString("url").String(),
					CoverURL: it.RawGetString("cover_url").String(),
					Info:     it.RawGetString("info").String(),
				})
			}
		})
	}
	return res, nil
}

func (s *LuaPlugin) Search(q string) ([]models.SearchItem, error) {
	return s.callSearchFunc("search", lua.LString(q))
}
func (s *LuaPlugin) GetPopular(p int) ([]models.SearchItem, error) {
	return s.callSearchFunc("get_popular", lua.LNumber(p))
}
func (s *LuaPlugin) GetLatest(p int) ([]models.SearchItem, error) {
	return s.callSearchFunc("get_latest", lua.LNumber(p))
}

func (s *LuaPlugin) GetDocument(u string) (models.Document, error) {
	pName := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	if !s.HasCapability("source") {
		fmt.Printf("[Security] [%s] Blocked unauthorized source call: get_document (Capability 'source' not enabled)\n", pName)
		return models.Document{}, nil
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.L.SetContext(ctx)

	if err := s.L.CallByParam(lua.P{Fn: s.L.GetGlobal("get_document"), NRet: 1, Protect: true}, lua.LString(u)); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			pName := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
			fmt.Printf("[Security] [%s] Execution timed out in get_document\n", pName)
		}
		return models.Document{}, err
	}
	ret := s.L.Get(-1)
	s.L.Pop(1)
	if tbl, ok := ret.(*lua.LTable); ok {
		luaStr := func(key string) string {
			v := tbl.RawGetString(key)
			if v.Type() == lua.LTNil {
				return ""
			}
			return v.String()
		}

		doc := models.Document{
			Title:    luaStr("title"),
			URL:      luaStr("url"),
			CoverURL: luaStr("cover_url"),
			Author:   luaStr("author"),
			Synopsis: luaStr("synopsis"),
			Genres:   luaStr("genres"),
			Status:   luaStr("status"),
			Chapters: []models.Chapter{},
		}
		if doc.Title == "" {
			doc.Title = "UNKNOWN"
		}
		chTbl := tbl.RawGetString("chapters")
		if chs, ok := chTbl.(*lua.LTable); ok {
			chs.ForEach(func(k, v lua.LValue) {
				if c, ok := v.(*lua.LTable); ok {
					doc.Chapters = append(doc.Chapters, models.Chapter{
						Title:  c.RawGetString("title").String(),
						URL:    c.RawGetString("url").String(),
						Order:  int(lua.LVAsNumber(c.RawGetString("id"))),
						Status: "pending",
					})
				}
			})
		}
		return doc, nil
	}
	return models.Document{}, fmt.Errorf("invalid return")
}

func (s *LuaPlugin) GetChapter(u string) (models.Chapter, error) {
	pName := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
	if !s.HasCapability("source") {
		fmt.Printf("[Security] [%s] Blocked unauthorized source call: get_chapter (Capability 'source' not enabled)\n", pName)
		return models.Chapter{}, nil
	}

	s.Mu.Lock()
	defer s.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.L.SetContext(ctx)

	if err := s.L.CallByParam(lua.P{Fn: s.L.GetGlobal("get_chapter"), NRet: 1, Protect: true}, lua.LString(u)); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			pName := strings.TrimSuffix(filepath.Base(s.Path), ".lua")
			fmt.Printf("[Security] [%s] Execution timed out in get_chapter\n", pName)
		}
		return models.Chapter{}, err
	}
	ret := s.L.Get(-1)
	s.L.Pop(1)
	if tbl, ok := ret.(*lua.LTable); ok {
		content := tbl.RawGetString("content").String()
		title := tbl.RawGetString("title").String()
		return models.Chapter{
			Title:   title,
			Content: CleanHTML(content, title),
			URL:     u,
		}, nil
	}
	return models.Chapter{}, fmt.Errorf("invalid return")
}

func (s *LuaPlugin) CleanHTML(html string, chapterTitle string) string {
	return CleanHTML(html, chapterTitle)
}
