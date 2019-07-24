package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
)

var templates *template.Template
var categories map[string][]categoriesData

type layoutData struct {
	Content string
	Data    interface{}
}

type headData struct {
	Title       string
	Description string
	Keywords    string
	Lang		string
}

type categoriesData struct {
	Name  string
	Slug  string
	Order int
	Lang  string
    headData
}

type articleData struct {
	Categories []categoriesData
	headData
}

type articleListData struct {
	Categories []categoriesData
	headData
}

type mainPageData struct {
	Categories []categoriesData
	headData
}

type noFoundData struct {
	Categories []categoriesData
	headData
}

func getLangFromRequest(r *http.Request) string {
	lang := r.Header.Get("Accept-Language")
	if lang == "" {
		lang = "pl"
	}
	return lang
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLangFromRequest(r);
	var data mainPageData
	data.Categories = categories[lang];
	data.Lang = lang
	renderWithLayout(w, "homepage.gohtml", data)
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLangFromRequest(r);
	var data articleData
	data.Categories = categories[lang];
	data.Lang = lang
	renderWithLayout(w, "article.gohtml", data)
}

func articleListHandler(w http.ResponseWriter, r *http.Request) {
	lang := getLangFromRequest(r);
	vars := mux.Vars(r)
	categorySlug := vars["category"]

	var cat *categoriesData

	for _, category := range categories[lang] {
		if category.Slug == categorySlug {
			cat = &category
			break
		}
	}

	if cat == nil {
		notFoundHandler(w,r)
	} else {
		var data articleListData
		data.Categories = categories[lang];
		data.Lang = lang
		data.Description = cat.Description
		data.Keywords = cat.Keywords
		data.Title = cat.Title
		renderWithLayout(w, "article-list.gohtml", data)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	lang := getLangFromRequest(r);
	var data noFoundData
	data.Categories = categories[lang];
	data.Lang = lang
	renderWithLayout(w, "404.gohtml", data)
}

func renderWithLayout(w http.ResponseWriter, name string, data interface{}) {
	buf := new(bytes.Buffer)

	if err := templates.ExecuteTemplate(buf, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := templates.ExecuteTemplate(w, "layout.gohtml", layoutData{Content: buf.String(), Data: data}); err != nil {
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

		if err := json.Unmarshal(jsonData, &category); err != nil {
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
	r.HandleFunc("/{lang:[a-z]{2}}/{category:[a-z\\-]+}", articleListHandler)
	r.HandleFunc("/{lang:[a-z]{2}}/{category:[a-z]+}/{slug:[a-z\\-]+}", articleHandler)
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
