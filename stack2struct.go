// This parses raw golang stack traces ([]byte) to a slice of well formated structs.

// As this package will need to evolve with the development of go's stack trace
// format, this is the stack format the package currently works with:
//
//  1: goroutine x [running]:                           <-- ignore this line
//  2: path/to/package.functionName()
//  3: path/to/responsible/file:lineNumber +0xdeadbeef  <-- memory address optional
//  ... repeat 2 + 3 for each stack trace element
//
// To work with this, you need a type satisfying the interface
//
//  type stackTrace interface {
//    AddEntry(lineNumber int, packageName string, fileName string, methodName string)
//  }
//
// and from there you can do whatever you like with the accumulated data.

package raygun4go

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// stackTrace is the interface a target stack has to satisfy.
type stackTrace interface {
	AddEntry(lineNumber int, packageName string, fileName string, methodName string)
}

// Current loads the current stacktrace into a given stack
func Current(stack stackTrace) {
	rawStack := make([]byte, 1<<16)
	rawStack = rawStack[:runtime.Stack(rawStack, false)]
	Parse(rawStack, stack)
}

// Parse loads the stack trace (given as trace) into the given stack.
// See Current() on how to obtain a stack trace.
func Parse(trace []byte, stack stackTrace) {
	lines := strings.Split(string(trace), "\n")

	var lineNumber int
	var fileName, packageName, methodName string

	for index, line := range lines[1:] {
		if len(line) == 0 {
			continue
		}
		if index%2 == 0 {
			packageName, methodName = extractPackageName(line)
		} else {
			lineNumber, fileName = extractLineNumberAndFile(line)
			stack.AddEntry(lineNumber, packageName, fileName, methodName)
		}
	}
}

// extractPageName receives a trace line and extracts packageName and
// methodName.
func extractPackageName(line string) (packageName, methodName string) {
	packagePath, packageNameAndFunction := splitAtLastSlash(line)
	parts := strings.Split(packageNameAndFunction, ".")
	
	if len(parts) > 1 {
		packageName = parts[0]
		if len(packagePath) > 0 {
			packageName = fmt.Sprintf("%s/%s", packagePath, packageName)
		}
		methodName = strings.Join(parts[1:], ".")
	} else {
		methodName = parts[0]
	}
	
	return
}

// extractLineNumberAndFile receives a trace line and extracts lineNumber and
// fileName.
func extractLineNumberAndFile(line string) (lineNumber int, fileName string) {
	_, fileAndLine := splitAtLastSlash(line)
	fileAndLine = removeSpaceAndSuffix(fileAndLine)
	parts := strings.Split(fileAndLine, ":")

    lineNumber = 0

    if len(parts) >= 2 {
	    numberAsString := parts[1]
	    number, _ := strconv.ParseUint(numberAsString, 10, 32)
	    lineNumber = int(number)
    }

	fileName = parts[0]
	return lineNumber, fileName
}

// splitAtLastSlash splits a string at the last found slash and returns the
// respective strings left and right of the slash.
func splitAtLastSlash(line string) (left, right string) {
	parts := strings.Split(line, "/")
	right = parts[len(parts)-1]
	left = strings.Join(parts[:len(parts)-1], "/")
	return
}

// removeSpaceAndSuffix splits the given string at ' ' and cuts off the part
// found after the last space.
func removeSpaceAndSuffix(line string) string {
	parts := strings.Split(line, " ")
	if len(parts) <= 1 {
	    return line
	}
	return strings.Join(parts[:len(parts)-1], " ")
}
