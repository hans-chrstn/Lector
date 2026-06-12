package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	lua "github.com/yuin/gopher-lua"
)

func (h *API) PluginRPC(c *fiber.Ctx) error {
	name := strings.ToLower(c.Params("name"))
	method := c.Params("method")

	p, ok := h.Engine.Plugins[name]
	if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "Plugin not found"})
	}

	if p.L == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Plugin is in a degraded state due to load failure and cannot execute code"})
	}

	p.Mu.Lock()
	defer p.Mu.Unlock()

	exports := p.L.GetGlobal("exports")
	if exports.Type() != lua.LTTable {
		if method == "get_document_actions" {
			return c.JSON([]interface{}{})
		}
		return c.Status(403).JSON(fiber.Map{"error": fmt.Sprintf("Plugin %s does not export any functions (exports table not found)", name)})
	}

	fn := p.L.GetField(exports, method)

	if fn.Type() != lua.LTFunction {
		if method == "get_document_actions" {
			return c.JSON([]interface{}{})
		}
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("RPC method %s not found in plugin %s", method, name)})
	}

	body := string(c.Body())
	if body == "" {
		body = "{}"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p.L.SetContext(ctx)

	if err := p.L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lua.LString(body)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("RPC error: %v", err)})
	}

	ret := p.L.Get(-1)
	p.L.Pop(1)

	if str, ok := ret.(lua.LString); ok {
		c.Set("Content-Type", "application/json")
		return c.SendString(string(str))
	}

	if tbl, ok := ret.(*lua.LTable); ok {
		data := tableToMap(tbl)
		return c.JSON(data)
	}

	return c.JSON(fiber.Map{"status": "success", "info": "RPC executed"})
}

func tableToMap(tbl *lua.LTable) interface{} {
	if tbl.MaxN() > 0 {
		arr := []interface{}{}
		tbl.ForEach(func(k, v lua.LValue) {
			arr = append(arr, luaValueToInterface(v))
		})
		return arr
	}
	res := make(map[string]interface{})
	tbl.ForEach(func(k, v lua.LValue) {
		res[k.String()] = luaValueToInterface(v)
	})
	return res
}

func luaValueToInterface(v lua.LValue) interface{} {
	switch v.Type() {
	case lua.LTString:
		return v.String()
	case lua.LTNumber:
		return float64(v.(lua.LNumber))
	case lua.LTBool:
		return bool(v.(lua.LBool))
	case lua.LTTable:
		return tableToMap(v.(*lua.LTable))
	default:
		return nil
	}
}
