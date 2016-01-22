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
		c, _ := New("key", "app", "1.0", []string{})
		c.Silent(true)
		So(c.appName, ShouldEqual, "app")
		So(c.apiKey, ShouldEqual, "key")
		So(c.context.Version, ShouldEqual, "1.0")
		So(c.context.DefaultTags, ShouldHaveSameTypeAs, []string{})
		So(c.context.Identifier(), ShouldHaveSameTypeAs, uuid.New())

		Convey("#New", func() {
			c, err := New("", "test", "", nil)
			So(c, ShouldEqual, nil)
			So(err, ShouldNotEqual, nil)

			c, err = New("test", "", "", nil)
			So(c, ShouldEqual, nil)
			So(err, ShouldNotEqual, nil)

			c, err = New("test", "test", "", nil)
			So(c, ShouldHaveSameTypeAs, &Client{})
			So(err, ShouldEqual, nil)
		})

		Convey("#CreateErrorEntry", func() {
			c, _ := New("key", "application", "1.0.0", []string{"local"})
			testErr := errors.New("test")
			e := c.CreateErrorEntry(testErr)
			So(e.err, ShouldEqual, testErr)
			So(e.version, ShouldEqual, c.context.Version)
			So(e.tags, ShouldHaveLength, len(c.context.DefaultTags))
			So(e.tags, ShouldResemble, c.context.DefaultTags)
			So(e.identifier, ShouldEqual, c.context.identifier)
		})

		Convey("#HandleError", func() {

			c.apiKey = "key"
			c.context.Version = "goconvey"
			c.context.DefaultTags = []string{"golang", "test"}

			defer c.HandleError()
			panic("Test: See if this works with Raygun")
		})

		Convey("#SubmitError", func() {
			ts := raygunEndpointStub()
			defer ts.Close()
			raygunEndpoint = ts.URL
			c, _ := New("key", "application", "1.0.0", []string{"local"})
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
			e.Request(r)
			e.CustomData(map[string]string{"foo": "bar"})
			e.User("Test User")

			c.Silent(false)

			c.SubmitError(e)
		})
	})
}

func TestErrorEntry(t *testing.T) {
	Convey("ErrorEntry", t, func() {
		c, _ := New("key", "application", "1.0.0", []string{"local"})
		testErr := errors.New("test")
		e := c.CreateErrorEntry(testErr)

		Convey("#Request", func() {
			r := &http.Request{}
			e.Request(r)
			So(e.request, ShouldResemble, r)
		})

		Convey("#Tags", func() {
			t := []string{"foo", "bar"}
			e.AppendTags(t)
			So(e.tags, ShouldResemble, append(c.context.DefaultTags, t...))
		})

		Convey("#CustomData", func() {
			cd := "foo"
			e.CustomData(cd)
			So(e.customData, ShouldResemble, cd)
		})

		Convey("#User", func() {
			u := "user"
			e.User(u)
			So(e.user, ShouldResemble, u)
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
