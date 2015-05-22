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

func startServer() {
	http.HandleFunc("/", index)
	http.HandleFunc("/forecast", forecastHandler)
	http.ListenAndServe(":8080", nil)
}

type IndexPage struct {
	RouteChoices []string
	DefaultTime  string
}

const (
	HTML5_DATE_FORMAT = "2006-01-02T15:04"
)

func index(w http.ResponseWriter, r *http.Request) {
	files, _ := ioutil.ReadDir("data/gpx_11")
	choices := make([]string, len(files))
	for i, _ := range files {
		choices[i] = files[i].Name()
	}

	t, _ := template.ParseFiles("index.html")
	p := &IndexPage{
		RouteChoices: choices,
		DefaultTime:  time.Now().Format(HTML5_DATE_FORMAT),
	}

	t.Execute(w, p)
}

func forecastHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	startTime, err := time.Parse(HTML5_DATE_FORMAT, r.FormValue("startTime"))
	if startTime.IsZero() || err != nil {
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
