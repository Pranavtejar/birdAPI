package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
)

type Template struct {
	tmpl *template.Template
}

type Cache struct {
	data map[string]string
	sync.RWMutex
	order []string
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

func home(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func (cache *Cache) write(data map[string]string) {
	cache.Lock()
	defer cache.Unlock()

	var all []map[string]string

	f, _ := os.ReadFile("bird.json")
	if len(f) > 0 {
		json.Unmarshal(f, &all)
	}

	all = append(all, data)

	b, _ := json.MarshalIndent(all, "", "  ")
	os.WriteFile("bird.json", b, 0644)

	cache.order = append(cache.order, data["birdName"])

	if len(cache.order) > 5 {
		cache.order = cache.order[1:]
	}
}

func form(cache *Cache) echo.HandlerFunc {
	return func(c echo.Context) error {

		newEntry := map[string]string{
			"birdName": c.FormValue("birdName"),
			"location": c.FormValue("location"),
		}

		cache.write(newEntry)

		return c.NoContent(http.StatusOK)
	}
}

func get(c echo.Context) error {
	var all []map[string]string

	f, _ := os.ReadFile("bird.json")
	if len(f) > 0 {
		json.Unmarshal(f, &all)
	}

	if len(all) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "no birds found",
		})
	}
	return c.JSON(http.StatusOK, all[len(all)-1])
}


func main() {
	e := echo.New()

	e.Renderer = &Template{
		tmpl: template.Must(
			template.ParseGlob("templates/*.html"),
		),
	}

	cache := &Cache{
		data: make(map[string]string),
		order: []string{},
	}

	e.GET("/", home)
	e.GET("/api", get)
	e.POST("/", form(cache))

	e.Start(":8080")
}






