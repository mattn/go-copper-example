package main

import (
	"log"
	"net/http"

	"github.com/mattn/go-slim"
	"github.com/tusharsoni/copper"
	"github.com/tusharsoni/copper/chttp"
	"github.com/tusharsoni/copper/clogger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	Content string `json:"content"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("posts.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Post{})

	t, err := slim.ParseFile("view/index.slim")
	if err != nil {
		log.Fatal(err)
	}

	params := copper.HTTPAppParams{
		Logger: clogger.NewStdLogger(),
		Routes: []chttp.Route{
			{
				Path:    "/",
				Methods: []string{http.MethodGet},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					var posts []Post
					if err := db.Order("id desc").Find(&posts).Error; err != nil {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					if err := t.Execute(w, map[string]interface{}{"posts": posts}); err != nil {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
				}),
			},
			{
				Path:    "/add",
				Methods: []string{http.MethodPost},
				Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if err := db.Create(&Post{Content: req.PostFormValue("content")}).Error; err != nil {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						return
					}
					http.Redirect(w, req, "/", http.StatusFound)
				}),
			},
		},
	}

	copper.RunHTTPApp(params)
}
