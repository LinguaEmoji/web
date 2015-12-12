package main

import (
    "github.com/go-martini/martini"

    "github.com/martini-contrib/render"

    "net/http"
)

func main() {
    m := martini.Classic()
    m.Use(render.Renderer(render.Options {
        Directory: "templates",
        Layout: "layout",
    }))
    handlers(m)
    m.Run()
}

func handlers(m *martini.ClassicMartini) {
    m.Get("/", homePage)
}

func homePage(r *http.Request, w http.ResponseWriter, ren render.Render) {
    ren.HTML(200, "home", nil)
}
