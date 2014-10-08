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
	"crypto/sha512"
	_ "expvar"
	"fmt"
	"html/template"
	"io"
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
	MD5         []byte  // [16]byte
	SHA1        []byte  // [20]byte
	SHA224      []byte  // [28]byte
	SHA256      []byte  // [32]byte
	SHA512	    []byte  // [64]byte
}

// constants and variables:
var template_base_path = "/go/src/github.com/scusi/Md5Webserver/"
//var template_base_path = ""
var templates = template.Must(template.ParseFiles(template_base_path+"tmpl/upload.html", template_base_path+"tmpl/download.html"))

// shows the upload form
func upHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(template_base_path+"tmpl/upload.html")
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
	// create new Checksum handles
	md5 := md5.New()
	sha1 := sha1.New()
	sha224 := sha256.New224()
	sha256 := sha256.New()
	sha512 := sha512.New()
	// create a MultiWriter to write to all handles at once
	mw := io.MultiWriter(md5, sha1, sha224, sha256, sha512)
	mw.Write(slurp)
	// get checksums of uploaded file
	md5sum := md5.Sum(nil)
	shasum := sha1.Sum(nil)
	sha224sum := sha224.Sum(nil)
	sha256sum := sha256.Sum(nil)
	sha512sum := sha512.Sum(nil)
	// parse uploaded data into my FileObject
	myFileObj := file{mpHeader.Filename, // filename
		mpHeader.Header.Get("Content-Type"), // Content-Type
		int64(len(slurp)),                   // size of file in bytes
		slurp,                               // file content
		md5sum,                              // md5sum of file content
		shasum,                              // sha1sum of file content
		sha224sum,                           // sha224sum
		sha256sum,                           // sha256sum
		sha512sum,			     // sha512sum 
	}
	// Parse and execute template with my FileObject
	t, _ := template.ParseFiles(template_base_path+"tmpl/download.html")
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
	//http.ListenAndServe(":9090", nil)
	http.ListenAndServe(":80", nil)
}
