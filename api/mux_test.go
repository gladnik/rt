package api

import (
	"net/http/httptest"
	"fmt"
	"net/http"
	"testing"
	
	. "github.com/aandryashin/matchers"
	. "github.com/aandryashin/matchers/httpresp"
	"io/ioutil"
	"encoding/json"
)

var (
	srv  *httptest.Server
	exit chan bool
)

func init() {
	srv = httptest.NewServer(Mux(exit))
}

func apiUrl(path string) string {
	return fmt.Sprintf("%s%s", srv.URL, path)
}

func TestPing(t *testing.T) {
	rsp, err := http.Get(apiUrl("/ping"))

	AssertThat(t, err, Is{nil})
	AssertThat(t, rsp, Code{http.StatusOK})
	AssertThat(t, rsp.Body, Is{Not{nil}})

	var data map[string]string
	bt, readErr := ioutil.ReadAll(rsp.Body)
	AssertThat(t, readErr, Is{nil})
	jsonErr := json.Unmarshal(bt, &data)
	AssertThat(t, jsonErr, Is{nil})
	_, hasUptime := data["uptime"]
	AssertThat(t, hasUptime, Is{true})
}

