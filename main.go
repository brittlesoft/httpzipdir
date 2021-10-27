package main

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"log"

	"github.com/Masterminds/sprig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/bytes"
	"github.com/landlock-lsm/go-landlock/landlock"
	flag "github.com/spf13/pflag"
)

var VERSION string

type HttpExport struct {
	Root          string
	BaseURL       string
	AutoIndex     bool
	DirZip        bool
	indexTemplate *template.Template
}

func NewHttpExport(root, baseURL string, autoindex, dirzip bool) (he *HttpExport, err error) {
	if !autoindex && !dirzip {
		return he, fmt.Errorf("Invalid options: both autoindex and dirzip are off")
	}

	t := template.New("dirlist").Funcs(template.FuncMap{"PathJoin": path.Join})
	t = t.Funcs(sprig.FuncMap())
	t, err = t.Parse(dirlistTemplate)
	if err != nil {
		return nil, err
	}
	he = &HttpExport{root, baseURL, autoindex, dirzip, t}
	return he, err
}

func (he *HttpExport) HttpHandler(c echo.Context) (err error) {
	r := c.Request()

	// Resolve urlpath to a path under root
	urlpath := path.Clean(r.URL.Path)
	urlpathnoprefix := strings.TrimPrefix(urlpath, he.BaseURL)
	realreqpath := path.Join(he.Root, urlpathnoprefix)
	//log.Printf("prefix: %s urlpath: %s urlpathnoprefix: %s realreqpath: %s\n", he.BaseURL, urlpath, urlpathnoprefix, realreqpath)

	stat, err := os.Stat(realreqpath)
	switch {
	case err != nil:
		// ZipDir magic
		// The req file doesn't exist but ends with .zip
		// if the req file without the zip extension is a directory,
		// send the directory contents as a zipfile
		if he.DirZip && strings.HasSuffix(realreqpath, ".zip") {
			realreqpath = strings.TrimSuffix(realreqpath, ".zip")
			stat, err = os.Stat(realreqpath)
			if err != nil || !stat.IsDir() {
				return notfound(c)
			}

			return he.handleZipDir(c, realreqpath)
		}
		return notfound(c)

	case stat.Mode().IsRegular():
		he.serveFile(c, stat, realreqpath)
	case he.AutoIndex && stat.Mode().IsDir():
		// Make sure GETs on directories end with a salt otherwise
		// Parent Directory link won't work as expected.
		// can't use urlpath here, path.Clean strips trailing slashes
		if !strings.HasSuffix(r.URL.Path, "/") {
			return c.Redirect(http.StatusMovedPermanently, urlpath+"/")
		}

		return he.dirList(c, realreqpath)
	default:
		return notfound(c)
	}

	return err
}

func (he *HttpExport) dirList(c echo.Context, reqpath string) (err error) {
	dirents, err := os.ReadDir(reqpath)
	if err != nil {
		log.Printf("Failed to readdir %s: %s\n", reqpath, err)
		return err
	}

	// Lifted from echo/middleware/static.go
	res := c.Response()
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	data := struct {
		DirName string
		Files   []interface{}
	}{
		DirName: c.Request().URL.Path,
	}

	for _, f := range dirents {

		info, err := f.Info()
		if err != nil {
			log.Printf("Error calling stat on %s: %s\n", f.Name(), err)
			continue
		}

		if strings.HasPrefix(f.Name(), ".") || !info.Mode().IsRegular() && !info.Mode().IsDir() {
			// Skip hidden files and non file / directories (symlinks, devices, etc...)
			continue
		}

		data.Files = append(data.Files, struct {
			Name    string
			Dir     bool
			ModTime time.Time
			Size    string
		}{f.Name(), f.IsDir(), info.ModTime(), bytes.Format(info.Size())})
	}
	return he.indexTemplate.Execute(res, data)
}

func (he *HttpExport) serveFile(c echo.Context, info os.FileInfo, reqpath string) (err error) {
	if strings.HasPrefix(path.Base(reqpath), ".") {
		return notfound(c)
	}

	return c.File(reqpath)
}

func (he *HttpExport) handleZipDir(c echo.Context, reqpath string) (err error) {
	filename := path.Base(reqpath + ".zip")
	c.Response().Header().Set("Content-Type", "application/zip")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Response().WriteHeader(http.StatusOK)

	zw := zip.NewWriter(c.Response())
	defer zw.Close()

	// Walk directory.
	filepath.Walk(reqpath, func(p string, info os.FileInfo, err error) error {
		if strings.HasPrefix(path.Base(p), ".") || !info.Mode().IsRegular() {
			// skip dotfiles, dirs, symlinks, devices, etc...
			return nil
		}

		// Entries in the zipfile will be rooted below the requested dirname
		// e.g URL:/tmp/testdir.zip -> ZIP:testdir/file_a testdir/dir1/file_b
		relativep := strings.TrimPrefix(p, reqpath)
		relativep = strings.TrimLeft(relativep, "/")
		//log.Printf("relativep: %s\n", relativep)
		ze, err := zw.CreateHeader(&zip.FileHeader{Name: relativep, Method: 0, Modified: info.ModTime()})
		if err != nil {
			return fmt.Errorf("Failed for %s: %s", p, err)
		}
		file, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("Cannot open file %s: %s\n", p, err)
		}
		defer file.Close()

		_, err = io.Copy(ze, file)
		return err
	})
	return err
}

func notfound(c echo.Context) error {
	return c.String(http.StatusNotFound, "Not Found")
}

func SetupHandlers(e *echo.Echo, prefix2root *map[string]string) *echo.Echo {
	e.Any("*", notfound)
	for prefix, root := range *prefix2root {
		if !strings.HasPrefix(prefix, "/") {
			log.Fatalf("Invalid prefix, must start with a slash (/): %s\n", prefix)
		}

		// resolve root to an absolute path and make sure it exists
		root, err := filepath.Abs(root)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := os.Stat(root); err != nil {
			log.Fatal(err)
		}

		cp := path.Clean(prefix)
		export, err := NewHttpExport(root, cp, true, true)
		if err != nil {
			log.Fatal(err)
		}

		cpany := path.Join(cp, "*")
		log.Printf("Adding allowed prefix: %s -> %s\n", cpany, root)
		e.GET(cpany, export.HttpHandler)
		log.Printf("Adding allowed prefix: %s -> %s\n", cp, root)
		e.GET(cp, export.HttpHandler)
	}

	return e
}

func main() {
	prefix2root := flag.StringToString("allow", nil, "Add an allowed url prefix->docroot mapping. e.g /media/patate/url=/path/to/media/patate")
	listen := flag.String("listen", "127.0.0.1", "Listen address")
	port := flag.Int("port", 10666, "Listen port")
	version := flag.Bool("version", false, "Show version and exit")
	landlocked := flag.Bool("landlocked", true, "Failure to restrict access to docroots using landlock(7) is fatal")

	flag.Parse()

	if *version {
		fmt.Printf("%s\n", VERSION)
		os.Exit(0)
	}

	if len(*prefix2root) == 0 {
		flag.Usage()
		log.Fatal("No configuration. Set --allow")
	}

	if *port < 1 {
		flag.Usage()
		log.Fatalf("Invalid port: %d\n", *port)
	}

	roots := []string{}
	for _, root := range *prefix2root {
		roots = append(roots, root)
	}
	err := landlock.V1.RestrictPaths(
		landlock.RODirs(roots...),
	)
	if err != nil {
		msg := fmt.Sprintf("Failed landlock for %s: %s\n", roots, err)
		if *landlocked {
			log.Fatal(msg)
		} else {
			log.Printf(msg)
		}
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	e = SetupHandlers(e, prefix2root)

	log.Printf("Listening On: %s:%d\n", *listen, *port)
	e.Logger.Fatal(e.Start(*listen + ":" + strconv.Itoa(*port)))

}
