// md5websrv.go - webservice to upload file and show MD5 Sum of it.
// 
//  go build md5websrv.go
//  ./md5websrv
//  go to http://localhost:9090/up/
// 
package main

import(
    "log"
    "fmt"
    "net/http"
    "html/template"
    "net/http/httputil"
    "io/ioutil"
    "crypto/md5"
)

// define a FileObject
type file struct {
    Name string
    ContentType string
    Content []byte
    MD5 [16]byte
}

// constants and variables:
var templates = template.Must(template.ParseFiles("tmpl/upload.html", "tmpl/download.html"))

func upHandler(w http.ResponseWriter, r *http.Request) {
    t, _ := template.ParseFiles("tmpl/upload.html")
    p := ""
    t.Execute(w, p)
}

func doHandler(w http.ResponseWriter, r *http.Request) {
    // get multipart-file from request
    mpf, mpHeader, err := r.FormFile("file")
    if err != nil {
        log.Fatal(err)
    }
    // read-in uploaded file
    slurp, err := ioutil.ReadAll(mpf)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    // get md5sum of uploaded file
    md5sum := md5.Sum(slurp)
    // parse uploaded data into my FileObject
    myFileObj := file{ mpHeader.Filename, mpHeader.Header.Get("Content-Type"), slurp, md5sum }
    // Parse and execute template with my FileObject
    t, _ := template.ParseFiles("tmpl/download.html")
    t.Execute(w, myFileObj)
}

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
