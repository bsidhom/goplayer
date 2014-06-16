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
	bindAddr string
	musicRoot string
	//bindHost  = flag.String("host", "[::1]", "host to bind to")
	//bindPort  = flag.Int("port", 8080, "http listen address")
	//musicRoot = flag.String("root", ".", "music root")
)

func readFlags() {
	bindHost  := flag.String("host", "[::1]", "host to bind to")
	bindPort  := flag.Int("port", 8080, "http listen address")
	root := flag.String("root", ".", "music root")
	flag.Parse()
	bindAddr = fmt.Sprintf("%s:%d", *bindHost, *bindPort)
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		log.Fatal(err)
	}
	musicRoot = absRoot
}

func main() {
	//flag.Parse()
	readFlags()
	http.HandleFunc("/", Index)
	http.HandleFunc(filePrefix, File)
	log.Printf("music root: %q", musicRoot)
	log.Printf("starting server on %s", bindAddr)
	err := http.ListenAndServe(bindAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
	log.Printf("serving index to %s", r.RemoteAddr)
}

func File(w http.ResponseWriter, r *http.Request) {
	fn := filepath.Join(musicRoot, r.URL.Path[len(filePrefix):])
	fi, err := os.Stat(fn)
	log.Printf("serving request %q to %s\n", r.URL, r.RemoteAddr)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err.Error())
	} else if fi.IsDir() {
		serveDirectory(fn, w, r)
	} else {
		http.ServeFile(w, r, fn)
	}
}

func serveDirectory(fn string, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err, ok := recover().(error); ok {
			http.Error(w, "error", http.StatusInternalServerError)
			log.Println(err.Error())
		}
	}()
	d, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

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
	log.Printf("serving directory: %s", fn)
}
