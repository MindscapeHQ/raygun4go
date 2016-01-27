package raygun4go

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/pborman/uuid"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	Convey("Client", t, func() {
		c, _ := New("app", "key")
		c.Version("1.0").Tags([]string{})
		c.Silent(true)
		So(c.appName, ShouldEqual, "app")
		So(c.apiKey, ShouldEqual, "key")
		So(c.context.Version, ShouldEqual, "1.0")
		So(c.context.Tags, ShouldHaveSameTypeAs, []string{})
		So(c.context.Identifier(), ShouldHaveSameTypeAs, uuid.New())

		Convey("#New", func() {
			c, err := New("", "test")
			So(c, ShouldEqual, nil)
			So(err, ShouldNotEqual, nil)

			c, err = New("test", "")
			So(c, ShouldEqual, nil)
			So(err, ShouldNotEqual, nil)

			c, err = New("test", "test")
			So(c, ShouldHaveSameTypeAs, &Client{})
			So(err, ShouldEqual, nil)
		})

		Convey("#Version", func() {
			v := "version"
			c.Version(v)
			So(c.context.Version, ShouldResemble, v)
		})

		Convey("#Tags", func() {
			t := []string{"foo", "bar"}
			c.Tags(t)
			So(c.context.Tags, ShouldResemble, t)
		})

		Convey("#CreateErrorEntry", func() {
			c, _ := New("application", "key")
			c.Version("1.0.0").Tags([]string{"local"})
			testErr := errors.New("test")
			e := c.CreateErrorEntry(testErr)
			So(e.Err, ShouldEqual, testErr)
			So(e.version, ShouldEqual, c.context.Version)
			So(e.tags, ShouldHaveLength, len(c.context.Tags))
			So(e.tags, ShouldResemble, c.context.Tags)
			So(e.identifier, ShouldEqual, c.context.identifier)
		})

		Convey("#HandleError", func() {

			c.apiKey = "key"
			c.context.Version = "goconvey"
			c.context.Tags = []string{"golang", "test"}

			defer c.HandleError()
			panic("Test: See if this works with Raygun")
		})

		Convey("#SubmitError", func() {
			ts := raygunEndpointStub()
			defer ts.Close()
			raygunEndpoint = ts.URL
			c, _ := New("application", "key")
			c.Version("1.0.0").Tags([]string{"local"})
			testErr := errors.New("test")
			e := c.CreateErrorEntry(testErr)

			u := "http://www.example.com?foo=bar&fizz[]=buzz&fizz[]=buzz2"
			r, _ := http.NewRequest("GET", u, nil)
			r.RemoteAddr = "1.2.3.4"
			r.PostForm = url.Values{
				"foo":  []string{"bar"},
				"fizz": []string{"buzz", "buzz2"},
			}
			r.Header.Add("Cookie", "cookie1=value1; cookie2=value2")
			e.SetRequest(r).
				SetCustomData(map[string]string{"foo": "bar"}).
				SetUser("Test User")

			c.Silent(false)

			c.SubmitError(e)
		})
	})
}

func TestErrorEntry(t *testing.T) {
	Convey("ErrorEntry", t, func() {
		c, _ := New("key", "application")
		c.Version("1.0.0").Tags([]string{"local"})
		testErr := errors.New("test")
		e := c.CreateErrorEntry(testErr)

		Convey("#Request", func() {
			r, _ := http.NewRequest("GET", "/", nil)
			e.SetRequest(r)
			So(e.Request.HTTPMethod, ShouldEqual, r.Method)
			So(e.Request.URL, ShouldEqual, r.URL.String())
		})

		Convey("#CustomData", func() {
			cd := "foo"
			e.SetCustomData(cd)
			So(e.CustomData, ShouldResemble, cd)
		})

		Convey("#User", func() {
			u := "user"
			e.SetUser(u)
			So(e.User, ShouldResemble, u)
		})

	})
}

func raygunEndpointStub() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if req.Method != "POST" || req.RequestURI != "/entries" {
				fmt.Println("raygunEndpointStub: URI not implemented")
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			// 403 Invalid API Key
			// The value specified in the header X-ApiKey did not match with a user.
			if req.Header.Get("X-ApiKey") == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			// 202 OK - Message accepted.
			w.WriteHeader(http.StatusAccepted)
		}))
}
