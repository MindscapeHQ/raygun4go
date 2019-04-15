package raygun4go

import (
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

		Convey("#Clone", func() {
			t := []string{"Critical", "Urgent", "Fix it now!"}
			c.Tags(t)

			cd := "foo"
			c.CustomData(cd)

			r := &http.Request{}
			c.Request(r)

			v := "1.2.3"
			c.Version(v)

			u := "user"
			c.User(u)

			clone := c.Clone()

			So(clone.appName, ShouldResemble, c.appName)
			So(clone.apiKey, ShouldResemble, c.apiKey)
			So(clone.silent, ShouldResemble, c.silent)
			So(clone.logToStdOut, ShouldResemble, c.logToStdOut)
			So(clone.asynchronous, ShouldResemble, c.asynchronous)
			So(clone.context.Request, ShouldResemble, c.context.Request)
			So(clone.context.Version, ShouldResemble, c.context.Version)
			So(clone.context.Tags, ShouldResemble, c.context.Tags)
			So(clone.context.CustomData, ShouldResemble, c.context.CustomData)
			So(clone.context.User, ShouldResemble, c.context.User)
			So(clone.context.identifier, ShouldResemble, c.context.identifier)
			So(clone.context.GetCustomGroupingKey, ShouldResemble, c.context.GetCustomGroupingKey)

			// After cloning, make some changes to the original client
			// to assert that they aren't picked up in the clone
			c.Tags([]string{"Expected"})
			c.CustomData("bar")
			newRequest, _ := http.NewRequest("POST", "https://my.api.io", nil)
			c.Request(newRequest)
			c.Version("2.3.4")
			c.User("user2")
			c.Silent(true)
			c.LogToStdOut(true)
			c.Asynchronous(true)
			c.CustomGroupingKeyFunction(func(error, PostData)string{return "customGroupingKey"})

			So(clone.silent, ShouldNotResemble, c.silent)
			So(clone.logToStdOut, ShouldNotResemble, c.logToStdOut)
			So(clone.asynchronous, ShouldNotResemble, c.asynchronous)
			So(clone.context.Request, ShouldNotResemble, c.context.Request)
			So(clone.context.Version, ShouldNotResemble, c.context.Version)
			So(clone.context.Tags, ShouldNotResemble, c.context.Tags)
			So(clone.context.CustomData, ShouldNotResemble, c.context.CustomData)
			So(clone.context.User, ShouldNotResemble, c.context.User)
			So(clone.context.GetCustomGroupingKey, ShouldNotResemble, c.context.GetCustomGroupingKey)
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

		Convey("#Silent", func() {
			So(c.silent, ShouldBeFalse)
			c.Silent(true)
			So(c.silent, ShouldBeTrue)
		})
		
		Convey("#CustomGroupingKeyFunction", func() {
			So(c.context.GetCustomGroupingKey, ShouldEqual, nil)
		    c.CustomGroupingKeyFunction(func(error, PostData)string{return "customGroupingKey"})
			So(c.context.GetCustomGroupingKey, ShouldNotEqual, nil)
		})

		Convey("#LogToStdOut", func() {
			So(c.logToStdOut, ShouldBeFalse)
			c.LogToStdOut(true)
			So(c.logToStdOut, ShouldBeTrue)
		})

		Convey("#Asynchronous", func() {
			So(c.asynchronous, ShouldBeFalse)
			c.Asynchronous(true)
			So(c.asynchronous, ShouldBeTrue)
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
			c.Silent(true)
			c.Request(r)
			c.apiKey = "key"
			c.CustomGroupingKeyFunction(func(error, PostData)string{return "customGroupingKey"})
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
