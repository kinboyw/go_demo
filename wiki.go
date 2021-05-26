package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
  "errors"
)

type Page struct{
  Title string
  Body []byte
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

var validPath = regexp.MustCompile("^/(edit|view|save)/([a-zA-Z0-9]+)$")

func getTitle(w http.ResponseWriter, r *http.Request) (string,error){
  title := validPath.FindStringSubmatch(r.URL.Path)
  if title == nil {
    http.NotFound(w,r)
    return "", errors.New("invalid Page Title")
  }
  return title[2], nil
}

func (p *Page) save() error{
  filename := "data/"+p.Title + ".txt"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
  err := templates.ExecuteTemplate(w, "tmpl/"+tmpl+".html", p)
  if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func loadPage(title string) (*Page, error){
  filename := "data/"+title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  return &Page{Title: title, Body: body}, nil
}

func handler(w http.ResponseWriter, r *http.Request){
  fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string){
  p,err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string){
  p,err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w,"edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string){
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
  return func (w http.ResponseWriter, r *http.Request){
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil{
      http.NotFound(w,r)
      return
    }
    fn(w, r, m[2])
  }
}

func main(){
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  log.Fatal(http.ListenAndServe(":8080", nil))
}

