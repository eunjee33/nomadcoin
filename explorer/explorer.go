package explorer

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"github.com/eunjee33/nomadcoin/blockchain"
)

var templates *template.Template //모든 template을 관리하는 객체

const (
	templateDir string = "explorer/templates/"
)

type homeData struct {
	PageTitle string //맨 앞글자를 대문자로 써야 home.html에서 접근 가능
	Blocks []*blockchain.Block //!!! -> 대문자로유되어야 한다
}

func home (rw http.ResponseWriter, r *http.Request) {
	data := homeData{"Home", nil}
	templates.ExecuteTemplate(rw, "home", data)
}

func add (rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET": 
		templates.ExecuteTemplate(rw, "add", nil)
	case "POST":
		r.ParseForm()
		blockchain.Blockchain().AddBlock()
		http.Redirect(rw, r, "/", http.StatusPermanentRedirect)
	}
}

func Start(port int) {
	handler := http.NewServeMux() //모든 request에 application.json을 추가해주는 middleware을 만듦
	templates = template.Must(template.ParseGlob(templateDir + "pages/*.gohtml"))
	templates = template.Must(templates.ParseGlob(templateDir + "partials/*.gohtml"))
	handler.HandleFunc("/", home)
	handler.HandleFunc("/add", add)
	fmt.Printf("Listening on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler)) //우리가 만든 multiplexe (ur를 사용할거다!!
}