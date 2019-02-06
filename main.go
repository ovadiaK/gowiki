package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var pwd string
var templates *template.Template
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func main() {
	pwd, _ = os.Getwd()
	templates = template.Must(template.ParseGlob("templates/*.gohtml"))

	http.HandleFunc("/index/", welcomeHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/new/", newHandler)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Page struct {
	Title string
	Body  []byte
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}
func pagePath(title string) string {
	return filepath.Join(pwd, "texts", title+".txt")
}
func (p *Page) save() error {
	return ioutil.WriteFile(pagePath(p.Title), p.Body, 0600)
}
func loadPage(title string) (*Page, error) {
	body, err := ioutil.ReadFile(pagePath(title))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	if err := templates.ExecuteTemplate(w, tmpl+".gohtml", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	var titles []string
	infos, err := ioutil.ReadDir(filepath.Join(pwd, "texts"))
	fmt.Print("welcome")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	for _, info := range infos {
		if !info.IsDir() {
			titles = append(titles, info.Name()[:(len(info.Name())-len(".txt"))])
		}
	}
	fmt.Println(titles)
	if err := templates.ExecuteTemplate(w, "index.gohtml", titles); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	if err := p.save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
func newHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Redirect(w, r, "/index/", http.StatusFound)
	}
	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
}
