package raygun4go

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/kaeuferportal/stack2struct"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRequestData(t *testing.T) {
	Convey("#NewRequestData", t, func() {
		u := "http://www.example.com?foo=bar&fizz[]=buzz&fizz[]=buzz2"
		r, _ := http.NewRequest("GET", u, nil)

		Convey("empty if no request given", func() {
			d := newRequestData(nil)
			So(d, ShouldResemble, requestData{})
		})

		Convey("basic data", func() {
			r.RemoteAddr = "1.2.3.4"

			d := newRequestData(r)
			So(d.HostName, ShouldEqual, "www.example.com")
			So(d.URL, ShouldEqual, u)
			So(d.HTTPMethod, ShouldEqual, "GET")
			So(d.IPAddress, ShouldResemble, "1.2.3.4")
		})

		Convey("Form", func() {
			r.PostForm = url.Values{
				"foo":  []string{"bar"},
				"fizz": []string{"buzz", "buzz2"},
			}
			expected := map[string]string{
				"foo":  "bar",
				"fizz": "[buzz; buzz2]",
			}

			d := newRequestData(r)
			So(d.Form, ShouldResemble, expected)
		})

		Convey("QueryString", func() {
			expected := map[string]string{
				"foo":    "bar",
				"fizz[]": "[buzz; buzz2]",
			}

			d := newRequestData(r)
			So(d.QueryString, ShouldResemble, expected)
		})

		Convey("Headers", func() {
			r.Header = map[string][]string{
				"foo":  {"bar"},
				"fizz": {"buzz"},
			}
			expected := map[string]string{
				"foo":  "bar",
				"fizz": "buzz",
			}

			d := newRequestData(r)
			So(d.Headers, ShouldResemble, expected)
		})
	})
}

func TestErrorData(t *testing.T) {
	Convey("#NewErrorData", t, func() {
		trace, _ := ioutil.ReadFile("_fixtures/stack_trace")
		e := errors.New("test error")
		stack := make(stackTrace, 0, 0)
		stack2struct.Parse(trace, &stack)

		d := newErrorData(e, stack[3:])

		expected := stackTrace{
			stackTraceElement{11, "foo/package1", "filename1.go", "method1·001()"},
			stackTraceElement{22, "foo/package2", "filename2.go", "(*action).method2(0x208304420)"},
		}
		So(d.Message, ShouldEqual, "test error")
		So(d.StackTrace[0], ShouldResemble, expected[0])
		So(d.StackTrace[1], ShouldResemble, expected[1])
	})
}

func TestUser(t *testing.T) {
	Convey("has an exported identifier", t, func() {
		u := user{"test"}
		So(u.Identifier, ShouldEqual, "test")
	})
}

func TestContext(t *testing.T) {
	Convey("has an exported identifier", t, func() {
		c := context{"test"}
		So(c.Identifier, ShouldEqual, "test")
	})
}

func Test(t *testing.T) {
	Convey("", t, func() {
	})
}
