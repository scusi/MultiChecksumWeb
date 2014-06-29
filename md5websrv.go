// md5websrv.go - webservice to upload file and show MD5 Sum of it.
//
//  go build md5websrv.go
//  ./md5websrv
//  go to http://localhost:9090/up/
//
package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	_ "expvar"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

// define a FileObject
type file struct {
	Name        string
	ContentType string
	Size        int64
	Content     []byte
	MD5         [16]byte
	SHA1        [20]byte
	SHA224      [28]byte
	SHA256      [32]byte
}

// constants and variables:
var templates = template.Must(template.ParseFiles("tmpl/upload.html", "tmpl/download.html"))

// shows the upload form
func upHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("tmpl/upload.html")
	p := ""
	t.Execute(w, p)
}

// takes upload request and processes it
func doHandler(w http.ResponseWriter, r *http.Request) {
	// get multipart-file from request
	mpf, mpHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// read-in uploaded file
	slurp, err := ioutil.ReadAll(mpf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// get md5sum of uploaded file
	md5sum := md5.Sum(slurp)
	shasum := sha1.Sum(slurp)
	sha224sum := sha256.Sum224(slurp)
	sha256sum := sha256.Sum256(slurp)
	// parse uploaded data into my FileObject
	myFileObj := file{mpHeader.Filename, // filename
		mpHeader.Header.Get("Content-Type"), // Content-Type
		int64(len(slurp)),                   // size of file in bytes
		slurp,                               // file content
		md5sum,                              // md5sum of file content
		shasum,                              // sha1sum of file content
		sha224sum,                           // sha224sum
		sha256sum}                           // sha256sum
	// Parse and execute template with my FileObject
	t, _ := template.ParseFiles("tmpl/download.html")
	t.Execute(w, myFileObj)
}

// dump the incoming request
func reqDumper(w http.ResponseWriter, r *http.Request) {
	dumpedReq, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal("httputil.DumpRequestOut:", err)
	}
	log.Printf("%s\n", dumpedReq)
	fmt.Fprintf(w, "%s", dumpedReq)
}

func main() {
	http.HandleFunc("/", upHandler)
	http.HandleFunc("/up/", upHandler)
	http.HandleFunc("/do/", doHandler)
	http.HandleFunc("/dump/", reqDumper)
	http.ListenAndServe(":9090", nil)
}
