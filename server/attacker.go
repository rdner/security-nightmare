package server

import (
	"io"
	"log"
	"net/http"
	"os"
)

type Attacker interface {
	OpenerCase(http.ResponseWriter, *http.Request)
	CSRFCase(http.ResponseWriter, *http.Request)
	CSRFFormCase(http.ResponseWriter, *http.Request)
	Gotcha(http.ResponseWriter, *http.Request)
}

func NewAttacker() Attacker {
	return &attacker{}
}

type attacker struct{}

func (a *attacker) OpenerCase(w http.ResponseWriter, r *http.Request) {
	serveCase(w, "opener")
}

func (a *attacker) CSRFCase(w http.ResponseWriter, r *http.Request) {
	serveCase(w, "csrf")
}

func (a *attacker) CSRFFormCase(w http.ResponseWriter, r *http.Request) {
	serveCase(w, "csrf_form")
}

func (a *attacker) Gotcha(w http.ResponseWriter, r *http.Request) {
	serveCase(w, "gotcha")
}

func serveCase(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusOK)
	file, err := os.Open("./cases/" + name + ".html")
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()
	io.Copy(w, file)
}
