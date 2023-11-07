package raygun4go

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/pborman/uuid"

	. "github.com/smartystreets/goconvey/convey"
)

// This key is required to send error reports to the Raygun dashboard.
// This value is only relevant if integrationTest (below) is set to true.
var apiKey = "key"

// integrationTest determines the mode of testing.
// If set to true:
//   - Test exceptions will be sent to Raygun using the provided apiKey.
//
// If set to false:
//   - Test exceptions won't be sent to Raygun, instead their payloads will
//     be printed to the console for local validation.
var integrationTest = false

func TestClient(t *testing.T) {
	Convey("Client", t, func() {
		c, _ := New("app", apiKey)
		So(c.appName, ShouldEqual, "app")
		So(c.apiKey, ShouldEqual, apiKey)
		So(c.context.Request, ShouldBeNil)
		So(c.context.Identifier(), ShouldHaveSameTypeAs, uuid.New())

		Convey("#New", func() {
			c, err := New("", "test")
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)

			c, err = New("test", "")
			So(c, ShouldBeNil)
			So(err, ShouldNotBeNil)

			c, err = New("test", "test")
			So(c, ShouldNotBeNil)
			So(c, ShouldHaveSameTypeAs, &Client{})
			So(err, ShouldBeNil)
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

			// After cloning, make some changes to the original client to assert that they aren't picked up in the clone
			c.Tags([]string{"Expected"})
			c.CustomData("bar")
			newRequest, _ := http.NewRequest("POST", "https://my.api.io", nil)
			c.Request(newRequest)
			c.Version("2.3.4")
			c.User("user2")
			c.Silent(true)
			c.LogToStdOut(true)
			c.Asynchronous(true)
			c.CustomGroupingKeyFunction(func(error, PostData) string { return "customGroupingKey" })

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
			So(c.context.GetCustomGroupingKey, ShouldBeNil)
			c.CustomGroupingKeyFunction(func(error, PostData) string { return "customGroupingKey" })
			So(c.context.GetCustomGroupingKey, ShouldNotBeNil)
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
			c.Silent(!integrationTest)
			c.Request(r)
			c.CustomGroupingKeyFunction(func(error, PostData) string { return "customGroupingKey" })
			c.context.Version = "goconvey"
			c.context.Tags = []string{"golang", "test"}
			c.context.CustomData = map[string]string{"foo": "bar"}
			c.context.User = "Test User"

			defer c.HandleError()
			panic("Test unhandled error")
		})

		Convey("#CreateError", func() {
			c.Silent(!integrationTest)
			c.context.Version = "goconvey"
			c.context.Tags = []string{"golang", "test"}
			c.context.CustomData = map[string]string{"foo": "bar"}
			c.context.User = "Test User"

			err := c.CreateError("Test CreateError")
			So(err, ShouldBeNil)
		})

		Convey("#CreateErrorWithStackTrace", func() {
			c.Silent(!integrationTest)
			c.context.Version = "goconvey"
			c.context.Tags = []string{"golang", "test"}
			c.context.CustomData = map[string]string{"foo": "bar"}
			c.context.User = "Test User"

			var customST StackTrace
			customST.AddEntry(42, "packageName", "fileName.go", "MethodName")

			err := c.CreateErrorWithStackTrace("Test CreateErrorWithStackTrace", customST)
			So(err, ShouldBeNil)
		})

		Convey("After testing", func() {
			fmt.Println()
			fmt.Println("==================================================================")
			if integrationTest {
				fmt.Println("Please check your Raygun dashboard to validate the three errors. You should expect:")
			} else {
				fmt.Println("Please validate the three error payloads above. You should expect:")
			}
			fmt.Println("1. An unhandled error with the following details:")
			fmt.Println("   - URL: http://www.example.com?foo=bar&fizz[]=buzz&fizz[]=buzz2")
			fmt.Println("   - Remote Address: 1.2.3.4")
			fmt.Println("   - Post Form: foo=bar, fizz=buzz, fizz=buzz2")
			fmt.Println("   - Headers: Cookie=cookie1=value1; cookie2=value2")
			fmt.Println("   - Custom Grouping Key: customGroupingKey")
			fmt.Println("   - Version: goconvey")
			fmt.Println("   - Tags: golang, test")
			fmt.Println("   - Custom Data: foo=bar")
			fmt.Println("   - User: Test User")
			fmt.Println("2. An error created with CreateError with the following details:")
			fmt.Println("   - Version: goconvey")
			fmt.Println("   - Tags: golang, test")
			fmt.Println("   - Custom Data: foo=bar")
			fmt.Println("   - User: Test User")
			fmt.Println("   - Error message: Test CreateError")
			fmt.Println("3. An error created with CreateErrorWithStackTrace with the following details:")
			fmt.Println("   - Version: goconvey")
			fmt.Println("   - Tags: golang, test")
			fmt.Println("   - Custom Data: foo=bar")
			fmt.Println("   - User: Test User")
			fmt.Println("   - Error message: Test CreateErrorWithStackTrace")
			fmt.Println("   - Custom Stack Trace: packageName.fileName.go:42 MethodName")
			fmt.Println("==================================================================")
		})
	})
}
