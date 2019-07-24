package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"encoding/json"
	"sort"
)

var templates *template.Template
var categories map[string][]categoriesData

type layoutData struct{
	Content string
	Data interface{}
}

type headData struct {
	Title string
	Description string
	Keywords string
}

type categoriesData struct {
	Name  string
	Slug  string
	Order int
	Lang  string
	headData
}

type mainPageData struct {
	Categories []categoriesData
	headData
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.Header.Get("Accept-Language")
	if lang == "" {
		lang = "pl"
	}
	var data mainPageData
	data.Categories = categories[lang]

	renderWithLayout(w, "homepage.gohtml", data)
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "article.gohtml", nil)
}

func articleListHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "article-list.gohtml", nil)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	renderWithLayout(w, "404.gohtml", nil)
}

func renderWithLayout(w http.ResponseWriter, name string, data interface{}) {
	buf := new(bytes.Buffer)

	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := templates.ExecuteTemplate(w, "layout.gohtml", layoutData{Content:buf.String(), Data: data}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func initData() {
	categories = make(map[string][]categoriesData)
	files, _ := filepath.Glob("data/*/*/info.json")
	for _, file := range files {
		var category categoriesData
		lang := path.Base(path.Dir(path.Dir(file)))

		jsonData, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		if err:=json.Unmarshal(jsonData, &category); err != nil {
			log.Fatal(err)
		}

		category.Lang = lang
		categories[lang] = append(categories[lang], category)
	}

	for _, items := range categories {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Order < items[j].Order
		})
	}
}


func main() {
	initData()
	templates = template.Must(template.ParseGlob("templates/**/*.gohtml"))

	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", homePageHandler)
	r.HandleFunc("/{Lang:[a-z]{2}}/{category:[a-z]+}", articleListHandler)
	r.HandleFunc("/{Lang:[a-z]{2}}/{category:[a-z]+}/{slug:[a-z\\-]+}", articleHandler)
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}