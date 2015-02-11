package raygun4go

import (
	"net/http"
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
			c.apiKey = "m/R9wT5Y4L9v4xj0aVAe8g=="
			c.context.Version = "goconvey"
			c.context.Tags = []string{"golang", "test"}
			c.context.CustomData = map[string]string{"foo": "bar"}
			c.context.User = "Jakob Test"
			defer c.HandleError()
			panic("Test: See if this works with Raygun")
		})
	})
}
