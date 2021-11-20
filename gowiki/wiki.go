package main

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"
)

type Page struct {
	Title string
	Body  []byte //byte is expected by io libs
}

var templates = template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// Recivers attatch the function to a type hence making them methods
func (p *Page) save() error {
	filename := p.Title + ".txt"

	// '0600' is octal representation for file permissions
	return os.WriteFile("./data/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	// the optional parameter is the error returned if any
	body, err := os.ReadFile("./data/" + filename)

	if err != nil {
		return nil, err // *Page is nil, error is in the err variable
	}

	// If no error then return nil for error field
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {

	p, err := loadPage(title)
	// if the request page doesnt exist then
	// redirect to the edit page
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {

	p, err := loadPage(title)
	// if the page doesnt exist
	// create a new page with given title
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {

	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}

	err := p.save()
	// handle the errors that might occur during the save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// save the new page and redirect
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// renders the template on screen
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	// first handle error in parsing
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// this is a closer function
		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2]) // m[2] contains the title
	}
}

func main() {

	// Create a page
	// p1 := &Page{Title: "TestPage", Body: []byte("This is a simple Page.")}
	// p1.save() // Save it to a file, NOTE the syntax, we are calling save() as if it is a member function of Page struct
	// // Its because thats exactly what it is
	// p2, _ := loadPage("TestPage") // load from a file
	// fmt.Println(string(p2.Body)) // Print it to console

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
