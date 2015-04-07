package raygun4go

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"code.google.com/p/go-uuid/uuid"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	Convey("Client", t, func() {
		c, _ := New("app", "key")
		c.Silent(true)
		So(c.appName, ShouldEqual, "app")
		So(c.apiKey, ShouldEqual, "key")
		So(c.context.Request, ShouldBeNil)
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

		Convey("#Request", func() {
			r := &http.Request{}
			c.Request(r)
			So(c.context.Request, ShouldResemble, r)
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

		Convey("#CustomData", func() {
			cd := "foo"
			c.CustomData(cd)
			So(c.context.CustomData, ShouldResemble, cd)
		})

		Convey("#User", func() {
			u := "user"
			c.User(u)
			So(c.context.User, ShouldResemble, u)
		})

		Convey("#HandleError", func() {
			u := "http://www.example.com?foo=bar&fizz[]=buzz&fizz[]=buzz2"
			r, _ := http.NewRequest("GET", u, nil)
			r.RemoteAddr = "1.2.3.4"
			r.PostForm = url.Values{
				"foo":  []string{"bar"},
				"fizz": []string{"buzz", "buzz2"},
			}
			r.Header.Add("Cookie", "cookie1=value1; cookie2=value2")
			c.Request(r)
			c.apiKey = "key"
			c.context.Version = "goconvey"
			c.context.Tags = []string{"golang", "test"}
			c.context.CustomData = map[string]string{"foo": "bar"}
			c.context.User = "Test User"
			defer c.HandleError()
			panic("Test: See if this works with Raygun")
		})

		Convey("#CreateError", func() {
			ts := raygunEndpointStub()
			defer ts.Close()
			raygunEndpoint = ts.URL
			c, _ := New("app", "key")
			c.Silent(false)
			c.apiKey = "key"
			c.CreateError("Test: See if this works with Raygun")
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
