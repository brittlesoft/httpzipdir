package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var defaultP2R = map[string]string{"/test": "testdata/"}
var testdirFilenames = []string{"directory/carottes", "directory/patates", "patates", "zipfile.zip"}

func TestGetInvalid(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	for _, p := range []string{
		"/",
		"/notfound",
		"/testdata",
		"/test.zip",
		"/test/testdir/.hidden",
		"/test/testdir/symlink"} {
		t.Run("GET "+p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			e.Router().Find(req.Method, req.URL.Path, c)
			c.Handler()(c)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}
}

func TestGetZip(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	req := httptest.NewRequest(http.MethodGet, "/test/testdir.zip", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	resp := rec.Result()

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/zip", resp.Header.Get("Content-Type"))
	assert.Equal(t, "attachment; filename=\"testdir.zip\"", resp.Header.Get("Content-Disposition"))

	body, _ := io.ReadAll(resp.Body)
	z, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	for i, f := range z.File {
		assert.Equal(t, testdirFilenames[i], f.Name)
	}

}

// dir with zipfile named like the directory (dir vs dir.zip)
// Requests for dir.zip should return the file and not the dynamically zipped directory
func TestGetDirWithZip(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	req := httptest.NewRequest(http.MethodGet, "/test/dir_with_zip/dir.zip", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	resp := rec.Result()

	body, _ := io.ReadAll(resp.Body)
	z, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	zipDirFiles := []string{"1", "2", "3", "4"}
	for i, f := range z.File {
		assert.Equal(t, zipDirFiles[i], f.Name)
	}
}

func TestGetDirNoSlash(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	req := httptest.NewRequest(http.MethodGet, "/test/testdir", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	resp := rec.Result()

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/test/testdir/", resp.Header.Get("Location"))
}

func TestGetDir(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	req := httptest.NewRequest(http.MethodGet, "/test/testdir/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetDotFile(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	req := httptest.NewRequest(http.MethodGet, "/test/testdir/.hidden", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetRegularFile(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &defaultP2R)
	req := httptest.NewRequest(http.MethodGet, "/test/testdir/patates", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Result().Body)
	assert.Equal(t, body, []byte("des patates\n"))
}

func TestInvalidOpts(t *testing.T) {
	e := echo.New()
	for _, opt := range []string{
		"",
		"invalid",
		"invalid1,invalid2",
		"noautoindex,invalid"} {
		t.Run("invalid opt: "+opt, func(t *testing.T) {
			_, _, err := SetupHandlers(e, &map[string]string{"/test": "testdata/:" + opt})
			assert.NotNil(t, err)
		})
	}
}

func TestNoAutoIndex(t *testing.T) {
	e := echo.New()
	e, _, _ = SetupHandlers(e, &map[string]string{"/test": "testdata/:noautoindex"})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/test/testdir/patates", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	assert.Equal(t, http.StatusOK, rec.Code)
	body, _ := io.ReadAll(rec.Result().Body)
	assert.Equal(t, body, []byte("des patates\n"))
}

func TestNoZipDir(t *testing.T) {
	e := echo.New()
	e, _, err := SetupHandlers(e, &map[string]string{"/test": "testdata/:nozipdir"})
	assert.Nil(t, err)
	req := httptest.NewRequest(http.MethodGet, "/test/testdir.zip", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNoZipDirNoAutoIndex(t *testing.T) {
	e := echo.New()
	e, _, err := SetupHandlers(e, &map[string]string{"/test": "testdata/:nozipdir,noautoindex"})
	assert.Nil(t, err)
	for url, expected := range map[string]int{
		"/test/testdir.zip":     http.StatusNotFound,
		"/test":                 http.StatusNotFound,
		"/test/testdir/patates": http.StatusOK} {
		t.Run(fmt.Sprintf("%s: %d", url, expected), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			e.Router().Find(req.Method, req.URL.Path, c)
			c.Handler()(c)
			assert.Equal(t, expected, rec.Code)
		})
	}
}

func TestIndexHtml(t *testing.T) {
	e := echo.New()
	e, _, err := SetupHandlers(e, &defaultP2R)
	assert.Nil(t, err)
	req := httptest.NewRequest(http.MethodGet, "/test/dir_with_index/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Router().Find(req.Method, req.URL.Path, c)
	c.Handler()(c)
	resp := rec.Result()

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "this is the index file\n", string(body))
}
