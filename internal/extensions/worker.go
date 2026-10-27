package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	lua "github.com/yuin/gopher-lua"
)

func Worker(in io.Reader, out io.Writer) error {
	if err := limitWorker(); err != nil {
		return err
	}
	var req Request
	if err := json.NewDecoder(io.LimitReader(in, 512<<10)).Decode(&req); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	result, err := execute(ctx, req)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(result)
}
func execute(ctx context.Context, req Request) (Result, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true, CallStackSize: 128, RegistrySize: 512, RegistryMaxSize: 8192, MinimizeStackMemory: true})
	defer L.Close()
	L.SetContext(ctx)
	for _, lib := range []struct {
		name string
		open lua.LGFunction
	}{{lua.BaseLibName, lua.OpenBase}, {lua.TabLibName, lua.OpenTable}, {lua.StringLibName, lua.OpenString}, {lua.MathLibName, lua.OpenMath}} {
		if err := L.CallByParam(lua.P{Fn: L.NewFunction(lib.open), Protect: true}, lua.LString(lib.name)); err != nil {
			return Result{}, err
		}
	}
	for _, name := range []string{"dofile", "loadfile", "load", "loadstring", "collectgarbage", "print", "module", "require", "getfenv", "setfenv", "newproxy"} {
		L.SetGlobal(name, lua.LNil)
	}
	api := L.NewTable()
	L.SetFuncs(api, map[string]lua.LGFunction{
		"fetch": func(L *lua.LState) int {
			body, err := fetch(ctx, L.CheckString(1), req.Hosts)
			if err != nil {
				L.RaiseError("request denied or failed: %v", err)
				return 0
			}
			L.Push(lua.LString(body))
			return 1
		},
		"json_decode": func(L *lua.LState) int {
			var v any
			if err := json.Unmarshal([]byte(L.CheckString(1)), &v); err != nil {
				L.RaiseError("invalid JSON")
				return 0
			}
			L.Push(toLua(L, v))
			return 1
		},
		"url_encode": func(L *lua.LState) int { L.Push(lua.LString(url.QueryEscape(L.CheckString(1)))); return 1 },
		"select": func(L *lua.LState) int {
			html, selector := L.CheckString(1), L.CheckString(2)
			if len(html) > 2<<20 || len(selector) > 500 {
				L.RaiseError("selection input too large")
				return 0
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				L.RaiseError("invalid document")
				return 0
			}
			result := L.NewTable()
			doc.Find(selector).EachWithBreak(func(index int, s *goquery.Selection) bool {
				if index >= 200 {
					return false
				}
				item := L.NewTable()
				item.RawSetString("text", lua.LString(s.Text()))
				href, _ := s.Attr("href")
				item.RawSetString("href", lua.LString(href))
				src, _ := s.Attr("src")
				item.RawSetString("src", lua.LString(src))
				result.Append(item)
				return true
			})
			L.Push(result)
			return 1
		},
	})
	L.SetGlobal("lector", api)
	if err := L.DoString(req.Source); err != nil {
		return Result{}, err
	}
	fn := L.GetGlobal(req.Action)
	if fn.Type() != lua.LTFunction {
		return Result{}, errors.New("action function not found")
	}
	if err := L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lua.LString(req.Input)); err != nil {
		return Result{}, err
	}
	value, err := fromLua(L.Get(-1), 0, map[*lua.LTable]bool{})
	if err != nil {
		return Result{}, err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return Result{}, err
	}
	if len(b) > 2<<20 {
		return Result{}, errors.New("result too large")
	}
	var result Result
	if err = json.Unmarshal(b, &result); err != nil {
		return Result{}, err
	}
	if result.Panels == nil {
		result.Panels = []Panel{}
	}
	if result.Entries == nil {
		result.Entries = []Entry{}
	}
	return result, result.validate(req.UI)
}
func toLua(L *lua.LState, v any) lua.LValue {
	switch v := v.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []any:
		t := L.NewTable()
		for _, x := range v {
			t.Append(toLua(L, x))
		}
		return t
	case map[string]any:
		t := L.NewTable()
		for k, x := range v {
			t.RawSetString(k, toLua(L, x))
		}
		return t
	}
	return lua.LNil
}
func fromLua(v lua.LValue, depth int, seen map[*lua.LTable]bool) (any, error) {
	if depth > 12 {
		return nil, errors.New("result nesting exceeds limit")
	}
	switch v := v.(type) {
	case *lua.LNilType:
		return nil, nil
	case lua.LBool:
		return bool(v), nil
	case lua.LNumber:
		return float64(v), nil
	case lua.LString:
		return string(v), nil
	case *lua.LTable:
		if seen[v] {
			return nil, errors.New("cyclic result")
		}
		seen[v] = true
		defer delete(seen, v)
		object := map[string]any{}
		array := []any{}
		var err error
		count := 0
		v.ForEach(func(k, x lua.LValue) {
			if err != nil {
				return
			}
			count++
			if count > 1000 {
				err = errors.New("too many result fields")
				return
			}
			var val any
			val, err = fromLua(x, depth+1, seen)
			if err != nil {
				return
			}
			if key, ok := k.(lua.LString); ok {
				object[string(key)] = val
			} else if n, ok := k.(lua.LNumber); ok && int(n) == len(array)+1 {
				array = append(array, val)
			} else {
				err = errors.New("invalid result key")
			}
		})
		if err != nil {
			return nil, err
		}
		if len(object) > 0 && len(array) > 0 {
			return nil, errors.New("mixed result table")
		}
		if len(object) > 0 {
			return object, nil
		}
		return array, nil
	default:
		return nil, errors.New("unsupported result value")
	}
}
