package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"
)

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func startServer() {
//	http.HandleFunc("/", http.File("index.html"))
	http.HandleFunc("/forecast", forecastHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// TODO: this exposes the entire bower_components subdir incl all source
	bower_fs := http.FileServer(http.Dir("bower_components"))
	http.Handle("/bower_components/", http.StripPrefix("/bower_components/", bower_fs))

	http.ListenAndServe(":8080", Log(http.DefaultServeMux))

}

type IndexPage struct {
	RouteChoices []string
	DefaultTime  string
}

// index renders a page with an HTML form for choosing the route, starting time and velocity.
func index(w http.ResponseWriter, r *http.Request) {
	files, _ := ioutil.ReadDir("data/gpx_11")
	choices := make([]string, len(files))
	for i, _ := range files {
		choices[i] = files[i].Name()
	}

	t, _ := template.ParseFiles("index.html")
//	t = t.Delims("[[", "]]")
	p := &IndexPage{
		RouteChoices: choices,
		DefaultTime:  time.Now().Format(time.RFC3339Nano),
	}

	t.Execute(w, p)
}

// forecastHandler handles a request to forecast a route at a specific time and place.
// startTime, velocity and route are parsed from FormValues (with reasonable defaults).
func forecastHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	startTime, err := time.Parse(time.RFC3339Nano, r.FormValue("startTime"))
	if startTime.IsZero() || err != nil {
		log.Printf("Got error %v parsing %v", err, r.FormValue("startTime"))
		startTime = time.Now()
	}

	mph, err := (strconv.ParseFloat(r.FormValue("velocity"), 32))
	var velocity = NewVelocityFromMph(mph)
	if velocity == 0 || err != nil {
		velocity = NewVelocityFromMph(10)
	}

	fname := r.FormValue("route")
	if fname == "" {
		fname = "bofax_alpine11.gpx"
	}
	track, err := ReadTrack(path.Join("data/gpx_11/", fname))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Track has %i points", track.Path().Length())
	fmt.Printf("modeling track starting at %s at velocity %.1f mph\n", startTime, velocity.Mph())
	track = ModelTrack(track, startTime, velocity)
	data := make([]ForecastedLocation, 0)
	for d := range ForecastTrack(track, sampleInterval) {
		d.Print()
		data = append(data, *d)
	}

	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
