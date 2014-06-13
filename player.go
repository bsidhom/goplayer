package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Entry struct {
	Name  string // name of the object
	IsDir bool
	Mode  os.FileMode
}

const (
	filePrefix = "/f/"
)

var (
	bindHost  = flag.String("host", "[::1]", "host to bind to")
	bindPort  = flag.Int("port", 8080, "http listen address")
	musicRoot = flag.String("root", "/home/flo/nfs/flo/Music/", "music root")
)

func main() {
	flag.Parse()
	http.HandleFunc("/", Index)
	http.HandleFunc(filePrefix, File)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *bindHost, *bindPort), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
	log.Printf("serving request to %s", r.RemoteAddr)
}

func File(w http.ResponseWriter, r *http.Request) {
	fn := filepath.Join(*musicRoot, r.URL.Path[len(filePrefix):])
	fi, err := os.Stat(fn)
	log.Print("File called: ", fn)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if fi.IsDir() {
		serveDirectory(fn, w, r)
		return
	}
	http.ServeFile(w, r, fn)
}

func serveDirectory(fn string, w http.ResponseWriter,
	r *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}()
	d, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	log.Print("serverDirectory called: ", fn)

	files, err := d.Readdir(-1)
	if err != nil {
		panic(err)
	}

	// Json Encode isn't working with the FileInfo interface,
	// therefore populate an Array of Entry and add the Name method
	entries := make([]Entry, len(files), len(files))

	for k := range files {
		//log.Print(files[k].Name())
		entries[k].Name = files[k].Name()
		entries[k].IsDir = files[k].IsDir()
		entries[k].Mode = files[k].Mode()
	}

	j := json.NewEncoder(w)

	if err := j.Encode(&entries); err != nil {
		panic(err)
	}
}
