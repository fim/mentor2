package main

import (
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "crypto/x509/pkix"
    "encoding/pem"
    "flag"
    "fmt"
    "html/template"
    "io"
    "log"
    "math/big"
    "net/http"
    "os"
    "path/filepath"
    "strconv"
    "time"
    "golang.org/x/crypto/bcrypt"
    "github.com/abbot/go-http-auth"
    //upnpl "github.com/NebulousLabs/go-upnp"
)

func usage() {
    fmt.Fprintf(os.Stderr, "usage %s [files]\n", os.Args[0])
    flag.PrintDefaults()
    os.Exit(2)
}

func generate_certs(crtfile string, keyfile string) error {

    priv, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return err
    }

    template := x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject: pkix.Name{
            Organization: []string{"Mentor"},
        },
        NotBefore: time.Now(),
        NotAfter:  time.Now().AddDate(0,0,10),

        KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
        BasicConstraintsValid: true,
    }

    derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
    if err != nil {
        return err
    }

    certOut, err := os.Create(crtfile)
    if err != nil {
        log.Fatalf("failed to open %s for writing: %s", crtfile, err)
    }
    pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
    certOut.Close()

    keyOut, err := os.OpenFile(keyfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Print("failed to open %s for writing:", keyfile, err)
        return err
    }
    pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
    keyOut.Close()

    return nil

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
        uploadLimit int, fileLists []string, ssl bool, password string) {
    if password != "" {
        authenticator := auth.NewBasicAuthenticator("Mentor", func(user, realm string) string {
            if user != "mentor" { return "" }
            hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
            return string(hash)
        })

        http.HandleFunc("/", auth.JustCheck(authenticator, s.IndexHandler))
        if upload {
            http.HandleFunc("/upload", auth.JustCheck(authenticator, s.UploadHandler))
            http.HandleFunc("/upload/", auth.JustCheck(authenticator, s.UploadHandler))
        }
    } else {
        http.HandleFunc("/", s.IndexHandler)

        if upload {
            http.HandleFunc("/upload", s.UploadHandler)
            http.HandleFunc("/upload/", s.UploadHandler)
        }
    }

    s.Files = make(map[string]struct{})
    for _, e := range fileLists {
        s.LoadPath(e)
    }
    if ! ssl {
        err := http.ListenAndServe(":" + strconv.Itoa(port), nil)
        if err != nil {
            log.Fatal("ListenAndServe: ", err)
        }
    } else {
        err := generate_certs("/tmp/crt", "/tmp/key")
        if err != nil {
            log.Fatal("generate_certs: ", err)
        }
        err = http.ListenAndServeTLS(":" + strconv.Itoa(port), "/tmp/crt", "/tmp/key", nil)
        if err != nil {
            log.Fatal("ListenAndServe: ", err)
        }
    }
}

func (s *MentorServer) UploadHandler (w http.ResponseWriter, r *http.Request) {
    log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
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
    log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
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
    ssl := flag.Bool("ssl", false, "Enable HTTPS (SSL)")
    //upnp := flag.Bool("upnp", false, "Enable UPnP hole-punching")
    upload := flag.Bool("upload", false, "Allow uploads through the service")
    uploadDir := flag.String("upload_dir", ".", "Upload directory")
    password := flag.String("password", "", "Password for accessing the service (username is mentor)")
    uploadLimit := flag.Int("upload_limit", 2, "Upload size limit (in MB)")
    flag.Parse()

    root := []string{}
    if len(flag.Args()) == 0 {
        log.Print("No paths were given. Using .")
        root = append(root, ".")
    } else {
        root = flag.Args()
    }

//    if *upnp {
//        d, _ := upnpl.Discover()
//        ip, _ := d.ExternalIP()
//        _ = d.Forward(uint16(*port), "Mentor")
//        log.Print("Listening on http://%s:%s", ip, *port)
//    }

    if (*password != "" && ! *ssl) {
        log.Print("It is recommended to enable SSL (-ssl flag) when using HTTP Basic auth")
    }

    var s MentorServer
    if *ssl {
        log.Print("Listening on https://0.0.0.0:", *port)
    } else {
        log.Print("Listening on http://0.0.0.0:", *port)
    }
    s.Start(*port, *upload, *uploadDir, *uploadLimit, root, *ssl, *password)
}
