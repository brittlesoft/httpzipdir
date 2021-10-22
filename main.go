package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	flag "github.com/spf13/pflag"
)

// TODO: support subdirs?

func notfound(c echo.Context) error {
	c.String(http.StatusNotFound, "Not Found")
	return nil
}

func makehandler(prefix, root string) func(echo.Context) error {

	handler := func(c echo.Context) error {
		r := c.Request()
		urlpath := path.Clean(r.URL.Path)

		if !strings.HasSuffix(urlpath, ".zip") {
			return notfound(c)
		}
		urlpath = strings.TrimSuffix(urlpath, ".zip")

		// Resolve urlpath to a path under root
		urlpathnoprefix := strings.TrimPrefix(urlpath, prefix)
		if len(urlpathnoprefix) == 0 || urlpathnoprefix == "/" {
			return notfound(c)
		}
		realreqpath := path.Join(root, urlpathnoprefix)
		//log.Printf("prefix: %s urlpath: %s urlpathnoprefix: %s realreqpath: %s\n", prefix, urlpath, urlpathnoprefix, realreqpath)

		stat, err := os.Stat(realreqpath)
		if err != nil || !stat.IsDir() {
			return notfound(c)
		}

		filename := path.Base(urlpath + ".zip")
		c.Response().Header().Set("Content-Type", "application/zip")
		c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		c.Response().WriteHeader(http.StatusOK)

		zw := zip.NewWriter(c.Response())
		defer zw.Close()

		// Walk directory.
		filepath.Walk(realreqpath, func(p string, info os.FileInfo, err error) error {
			if !info.Mode().IsRegular() {
				// skip dirs, synlinks, devices, etc...
				return nil
			}

			// Entries in the zipfile will be rooted below the requested dirname
			// e.g URL:/tmp/testdir.zip -> ZIP:testdir/file_a testdir/dir1/file_b
			relativep := strings.TrimPrefix(p, realreqpath)
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

		return nil
	}

	return handler
}

func SetupHandlers(e *echo.Echo, prefix2root *map[string]string) *echo.Echo {
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
		cpany := path.Join(cp, "*")
		log.Printf("Adding allowed prefix: %s -> %s\n", cpany, root)
		e.GET(cpany, makehandler(cp, root))
	}
	e.Any("*", notfound)

	return e
}

func main() {
	prefix2root := flag.StringToString("allow", nil, "Add an allowed url prefix->docroot mapping. e.g /media/patate/url=/path/to/media/patate")
	listen := flag.String("listen", "127.0.0.1", "Listen address")
	port := flag.Int("port", 10666, "Listen port")

	flag.Parse()

	if len(*prefix2root) == 0 {
		flag.Usage()
		log.Fatal("No configuration. Set --allow")
	}

	if *port < 1 {
		flag.Usage()
		log.Fatalf("Invalid port: %d\n", *port)
	}

	log.Printf("Listening On: %s:%d\n", *listen, *port)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e = SetupHandlers(e, prefix2root)
	e.Logger.Fatal(e.Start(*listen + ":" + strconv.Itoa(*port)))

}
