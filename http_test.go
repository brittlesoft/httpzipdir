package main

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var defaultP2R = map[string]string{"/test": "testdata/"}
var testdirFilenames = []string{"directory/carottes", "directory/patates", "patates"}

func TestGetInvalid(t *testing.T) {
	e := echo.New()
	e = SetupHandlers(e, &defaultP2R)
	for _, p := range []string{"/", "/notfound", "/test", "/testdata", "/test/testdir", "/test/.zip", "/test.zip"} {
		req := httptest.NewRequest(http.MethodGet, p, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		e.Router().Find(req.Method, req.URL.Path, c)
		c.Handler()(c)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	}
}

func TestGetZip(t *testing.T) {
	e := echo.New()
	e = SetupHandlers(e, &defaultP2R)
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
