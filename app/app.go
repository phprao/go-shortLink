package app

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/validator.v2"
	"log"
	"net/http"
	"shortLink/lib"
	"shortLink/middlewares"
)

type App struct {
	Router *mux.Router
	Middlewares *middlewares.Middleware
	Config *lib.Env
}

type shortenReq struct {
	URL                 string `json:"url" validate:"nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

type shortlinkResp struct {
	ShortLink string `json:"shortLink"`
}

func (a *App) Initialize() {
	// set log format
	// LstdFlags -> log date and time
	// Lshortfile -> log filename and line no
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// connect redis
	a.Config = lib.GetEnv()

	// set router object
	a.Router = mux.NewRouter()

	// add middleware
	a.Middlewares = &middlewares.Middleware{}

	// init router to handle
	a.initializeRouter()
}

func (a *App) initializeRouter() {
	// http://localhost:8080/api/shorten
	//a.Router.HandleFunc("/api/shorten", a.createShortLink).Methods("POST")
	// http://localhost:8080/api/info?shortLink=B
	//a.Router.HandleFunc("/api/info", a.getShortLinkInfo).Methods("GET")
	// http://localhost:8080/A
	//a.Router.HandleFunc("/{shortLink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")

	m := alice.New(a.Middlewares.LoggingHandler, a.Middlewares.RecoverHandler)

	// 优化，增加中间件
	a.Router.Handle("/api/shorten",
		m.ThenFunc(a.createShortLink)).Methods("POST")
	a.Router.Handle("/api/info",
		m.ThenFunc(a.getShortLinkInfo)).Methods("GET")
	a.Router.Handle("/{shortLink:[a-zA-Z0-9]{1,11}}",
		m.ThenFunc(a.redirect)).Methods("GET")
}

func (a *App) createShortLink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, lib.StatusError{Code: http.StatusBadRequest,
			Err: fmt.Errorf("parse parameters failed: %v", r.Body)})
		return
	}

	// 校验参数
	if err := validator.Validate(req); err != nil {
		respondWithError(w, lib.StatusError{Code: http.StatusBadRequest,
			Err: fmt.Errorf("validate parameters failed: %v", req)})
		return
	}
	defer r.Body.Close()

	str, err := a.Config.S.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		respondWithError(w, err)
	}else{
		respondWithJSON(w, http.StatusAlreadyReported, shortlinkResp{str})
	}
}

func (a *App) getShortLinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	s := vals.Get("shortLink")

	str, err := a.Config.S.ShortLinkInfo(s)
	if err != nil {
		respondWithError(w, err)
	}else{
		respondWithJSON(w, http.StatusAlreadyReported, str)
	}
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str, err := a.Config.S.UnShorten(vars["shortLink"])
	if err != nil {
		respondWithError(w, err)
	}else{
		http.Redirect(w, r, str, http.StatusTemporaryRedirect)
	}
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, err error){
	switch e := err.(type) {
	case lib.Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		respondWithJSON(w, e.Status(), e.Error())
	default:
		respondWithJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	resp, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)

	w.Write(resp)
}