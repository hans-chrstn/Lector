package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/core/interfaces"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
	"github.com/user/lector/internal/repository"
	"github.com/user/lector/internal/services"
)

func main() {
	db.InitDB("lector.db")
	os.MkdirAll("exports", 0755)
	os.MkdirAll("plugins", 0755)
	services.EnsureUploadsDir()

	uploadLimitMB, _ := strconv.Atoi(os.Getenv("MAX_UPLOAD_SIZE"))
	if uploadLimitMB <= 0 {
		uploadLimitMB = 100
	}

	app := fiber.New(fiber.Config{
		BodyLimit: uploadLimitMB * 1024 * 1024,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob: https:; style-src 'self' 'unsafe-inline' https://cdn.plyr.io; font-src 'self' data:; connect-src 'self' https://*; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://cdn.plyr.io; worker-src 'self' blob:; frame-ancestors 'self'; object-src 'none';")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	})

	app.Use(limiter.New(limiter.Config{
		Max:        600,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			return strings.HasPrefix(path, "/_app") ||
				strings.HasPrefix(path, "/api/proxy-image") ||
				strings.HasSuffix(path, ".js") ||
				strings.HasSuffix(path, ".css") ||
				strings.HasSuffix(path, ".svg") ||
				strings.HasSuffix(path, ".png") ||
				strings.HasSuffix(path, ".jpg") ||
				strings.HasSuffix(path, ".woff2")
		},
	}))

	authUser := os.Getenv("AUTH_USER")
	authPass := os.Getenv("AUTH_PASSWORD")
	if authUser != "" && authPass != "" {
		app.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				authUser: authPass,
			},
			Realm: "Lector Restricted Area",
		}))
	}

	heavyLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).SendString("Too many requests, please slow down.")
		},
	})
	app.Use("/api/upload", heavyLimiter)
	app.Use("/api/documents/*/export", heavyLimiter)

	origins := os.Getenv("CORS_ALLOW_ORIGINS")
	if origins != "" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: origins,
		}))
	}

	pluginStore := repository.NewPluginRepository()
	plugins := loadPlugins(pluginStore)
	plugin.GlobalPlugins = plugins

	engine := &plugin.PluginEngine{
		Store:   pluginStore,
		Plugins: plugins,
	}

	api.RegisterRoutes(app, engine)

	app.Static("/", "./public")
	app.Static("/uploads", "./uploads")
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("[Server] Lector starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func loadPlugins(store interfaces.PluginDataStore) map[string]*plugin.LuaPlugin {
	pluginsMap := make(map[string]*plugin.LuaPlugin)
	pluginDir := "plugins"
	os.MkdirAll(pluginDir, 0755)

	files, _ := os.ReadDir(pluginDir)
	for _, file := range files {
		var name string
		var path string

		if file.IsDir() {
			name = strings.ToLower(file.Name())
			path = filepath.Join(pluginDir, file.Name(), "init.lua")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				continue
			}
		} else if filepath.Ext(file.Name()) == ".lua" {
			name = strings.ToLower(file.Name()[:len(file.Name())-4])
			path = filepath.Join(pluginDir, file.Name())
			if _, exists := pluginsMap[name]; exists {
				continue
			}
		} else {
			continue
		}

		var p models.Plugin
		result := db.DB.Where("name = ?", name).First(&p)
		if result.Error != nil {
			p = models.Plugin{Name: name, IsEnabled: true}
			db.DB.Create(&p)
		}

		if p.IsEnabled {
			s, err := plugin.NewLuaPlugin(name, path, store)
			if err == nil {
				pluginsMap[name] = s
				log.Printf("[Plugin] Loaded: %s (Verified: %v)", name, s.IsVerified)
			} else {
				log.Printf("[Plugin] Failed to load %s from %s: %v", name, path, err)
			}
		} else {
			log.Printf("[Plugin] Skipping disabled plugin: %s", name)
		}
	}
	return pluginsMap
}
