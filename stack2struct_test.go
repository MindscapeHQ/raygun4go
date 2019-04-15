package raygun4go

import (
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testElement struct {
	lineNumber  int
	packageName string
	fileName    string
	methodName  string
}

type testStack []testElement

func (t *testStack) AddEntry(lineNumber int, packageName, fileName, methodName string) {
	*t = append(*t, testElement{lineNumber, packageName, fileName, methodName})
}

func TestStack2Struct(t *testing.T) {
	Convey("#splitAtLastSlash", t, func() {
		testLine := "foo/bar/baz"
		left, right := splitAtLastSlash(testLine)
		So(left, ShouldEqual, "foo/bar")
		So(right, ShouldEqual, "baz")
	})

	Convey("#removeSpaceAndSuffix", t, func() {
		testLine := "foo:bar baz"
		result := removeSpaceAndSuffix(testLine)
		So(result, ShouldEqual, "foo:bar")
	})

	Convey("#Parse", t, func() {
		buf, _ := ioutil.ReadFile("_fixtures/stack_trace")

		stack := make(testStack, 0, 0)
		Parse(buf, &stack)

		expected := testStack{
			testElement{13,
				"main",
				"stack2struct_test.go",
				"func.001()"},
			testElement{44,
				"github.com/smartystreets/goconvey/convey",
				"registration.go",
				"(*action).Invoke(0x208304420)"},
		}

		So(len(stack), ShouldEqual, 5)
		firstEntry := stack[0]
		So(firstEntry.lineNumber, ShouldEqual, expected[0].lineNumber)
		So(firstEntry.packageName, ShouldEqual, expected[0].packageName)
		So(firstEntry.fileName, ShouldEqual, expected[0].fileName)
		So(firstEntry.methodName, ShouldEqual, expected[0].methodName)

		So(stack[0], ShouldResemble, expected[0])
		So(stack[1], ShouldResemble, expected[1])
	})

}
