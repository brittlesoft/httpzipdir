package main

import (
	"archive/zip"
	"flag"
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
)

type DirZip struct {
	Prefix string
	Root   string
}

func (dz *DirZip) handler(c echo.Context) error {
	r := c.Request()
	urlpath := path.Clean(r.URL.Path)
	realreqpath := dz.Root + "/" + strings.TrimPrefix(urlpath, dz.Prefix)
	stat, err := os.Stat(realreqpath)
	if err != nil || !stat.IsDir() {
		c.NoContent(http.StatusNotFound)
		return nil
	}

	filename := path.Base(urlpath + ".zip")
	c.Response().Header().Set("Content-Type", "application/zip")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Response().WriteHeader(http.StatusOK)

	zw := zip.NewWriter(c.Response())
	defer zw.Close()
	// Walk directory.
	filepath.Walk(realreqpath, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// XXX add support for subdirectories,  also handle symlinks
			return nil
		}

		fmt.Printf("path: %s\n", p)

		// XXX with subdir support this will have to change
		ze, err := zw.CreateHeader(&zip.FileHeader{Name: path.Base(p), Method: 0, Modified: info.ModTime()})
		if err != nil {
			return fmt.Errorf("Failed for %p: %s", err)
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

func main() {
	var allowedPrefix, docRoot, listen string
	var port int
	flag.StringVar(&allowedPrefix, "allowedPrefix", "", "URL prefix for which request will be allowed")
	flag.StringVar(&docRoot, "docRoot", "", "Document root. Prefix is stripped from this and the rest is looked up under here")
	flag.StringVar(&listen, "listen", "127.0.0.1", "Listen address")
	flag.IntVar(&port, "port", 10666, "Listen port")

	flag.Parse()

	if docRoot == "" || allowedPrefix == "" {
		flag.CommandLine.Usage()
		log.Fatal("Missing docRoot or allowedPrefix")
	}

	if port < 1 {
		flag.CommandLine.Usage()
		log.Fatalf("Invalid port: %d\n", port)
	}

	dz := DirZip{Prefix: allowedPrefix, Root: docRoot}

	log.Printf("Root: %s\nListening On: %s:%d\n", docRoot, listen, port)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET(path.Clean(allowedPrefix+"/*"), dz.handler)
	e.Logger.Fatal(e.Start(listen + ":" + strconv.Itoa(port)))

}
