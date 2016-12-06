package main

import (
    "flag"
    "fmt"
    "html/template"
    "io"
    "log"
	"net/http"
    "os"
    "path/filepath"
    "strconv"
)

func usage() {
    fmt.Fprintf(os.Stderr, "usage %s [files]\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(2)
}

type MentorServer struct {
    Port    int
    Files   map[string]struct{}
}

func (s *MentorServer) LoadPath(root string) {
    filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
        if err != nil {
            return nil
        }
        if !f.IsDir() {
            absPath, _ := filepath.Abs(path)
            s.Files[absPath] = struct{}{}

        }
        return nil
    })
}

func (s *MentorServer) Start(port int, upload bool, uploadDir string,
        uploadLimit int, fileLists []string) {
    http.HandleFunc("/", s.IndexHandler)

    if upload {
        http.HandleFunc("/upload", s.UploadHandler)
        http.HandleFunc("/upload/", s.UploadHandler)
    }

    s.Files = make(map[string]struct{})
    for _, e := range fileLists {
        s.LoadPath(e)
    }
    err := http.ListenAndServe(":" + strconv.Itoa(port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func (s *MentorServer) UploadHandler (w http.ResponseWriter, r *http.Request) {

    log.Printf("Handling upload")
    log.Printf(r.Method)
    switch r.Method {
        case "GET":
            t := template.New("IndexPage")
            t, _ = t.Parse(`<h1>Mentor</h1>
                <pre>Mode: <a href="/">Download</a> | <a href="/upload">Upload</a></pre>
                <pre>#############################################################</pre>
                <script src="https://rawgit.com/fim/dropzone/master/dist/dropzone.js"></script>
                <link rel="stylesheet" href="https://rawgit.com/fim/dropzone/master/dist/dropzone.css">
                <form action="/upload" class="dropzone"></form>
            `)
            t.Execute(w, s.Files)
        case "POST":
            file, handler, err := r.FormFile("file")

            f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0644)
            if err != nil {
                fmt.Println(err)
                return
            }

            defer f.Close()
            io.Copy(f, file)

        default:
            http.Error(w, "Unsupported Method", 405)
    }
}

func (s *MentorServer) IndexHandler (w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        if _, ok := s.Files[r.URL.Path]; ok {
            http.ServeFile(w, r, r.URL.Path)
        } else if r.URL.Path == "/upload" {
            http.Error(w, "Uploading is not enabled", 404)
        } else {
            http.Error(w, "404 File not found", 404)
        }
        return
    }

    t := template.New("IndexPage")
    t, _ = t.Parse(`<h1>Mentor</h1>
        <pre>Mode: <a href="/">Download</a> | <a href="/upload">Upload</a></pre>
        <pre>#############################################################</pre>
        <table>
        {{ range $k, $v := . }}
            <tr><td>{{ $k }}</td><td><a href="{{ html $k }}">[GET]</a></td></tr>
        {{ end }}
        </table>
    `)
	t.Execute(w, s.Files)
}

func main() {
    port := flag.Int("port", 61234, "Port number")
    upload := flag.Bool("upload", false, "Allow uploads")
    uploadDir := flag.String("upload_dir", "/tmp", "Upload directory")
    uploadLimit := flag.Int("upload_limit", 2, "Upload size limit (in MB)")
    flag.Parse()

    root := []string{}
    if len(flag.Args()) == 0 {
        log.Print("No paths were given. Using .")
        root = append(root, ".")
    } else {
        root = flag.Args()
    }

    var s MentorServer
    s.Start(*port, *upload, *uploadDir, *uploadLimit, root)
}
