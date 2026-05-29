package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/user/lector/internal/api"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/plugin"
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

	app.Use(limiter.New(limiter.Config{
		Max:        120,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
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
	if origins == "" {
		origins = "*"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: origins,
	}))

	plugins := loadPlugins()
	plugin.GlobalPlugins = plugins

	api.RegisterRoutes(app, plugins)

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

func loadPlugins() map[string]*plugin.LuaPlugin {
	pluginsMap := make(map[string]*plugin.LuaPlugin)

	pluginDir := "plugins"
	os.MkdirAll(pluginDir, 0755)

	files, _ := os.ReadDir(pluginDir)

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".lua" {
			name := file.Name()[:len(file.Name())-4]

			var p models.Plugin
			result := db.DB.Where("name = ?", name).First(&p)
			if result.Error != nil {
				p = models.Plugin{Name: name, IsEnabled: true}
				db.DB.Create(&p)
			}

			if p.IsEnabled {
				s, err := plugin.NewLuaPlugin(filepath.Join(pluginDir, file.Name()))
				if err == nil {
					pluginsMap[name] = s
				} else {
					log.Printf("[Plugin] Failed to load %s: %v", name, err)
				}
			} else {
				log.Printf("[Plugin] Skipping disabled plugin: %s", name)
			}
		}
	}
	return pluginsMap
}
