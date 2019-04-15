package raygun4go

import (
	"net/http"
	"os"
	"time"
)

// postData is the outmost element of the Raygun-REST-API
type PostData struct {
	OccuredOn string      `json:"occurredOn"` // the time the error occured on, format 2006-01-02T15:04:05Z
	Details   DetailsData `json:"details"`    // all the details needed by the API
}

// newPostData triggers the creation of and returns a PostData-struct. It needs
// the configured context from the Client, the error and the corresponding
// stack trace.
func newPostData(context contextInformation, err error, stack StackTrace) PostData {
	return PostData{
		OccuredOn: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		Details:   newDetailsData(context, err, stack),
	}
}

// detailsData is the container holding all information regarding the more
// detailed circumstances the error occured in.
type DetailsData struct {
	MachineName    string         `json:"machineName"`    // the machine's hostname
	Version        string         `json:"version"`        // the version from context
	Error          ErrorData      `json:"error"`          // everything we know about the error itself
	Tags           []string       `json:"tags"`           // the tags from context
	UserCustomData UserCustomData `json:"userCustomData"` // the custom data from the context
	Request        RequestData    `json:"request"`        // the request from the context
	User           User           `json:"user"`           // the user from the context
	Context        Context        `json:"context"`        // the identifier from the context
	Client         ClientData     `json:"client"`         // information on this client
	GroupingKey    *string        `json:"groupingKey"`    // a custom key that Raygun will use for grouping errors
}

// newDetailsData returns a struct with all known details. It needs the context,
// the error and the stack trace.
func newDetailsData(c contextInformation, err error, stack StackTrace) DetailsData {
	hostname, e := os.Hostname()
	if e != nil {
		hostname = "not available"
	}

	return DetailsData{
		MachineName:    hostname,
		Version:        c.Version,
		Error:          newErrorData(err, stack),
		Tags:           c.Tags,
		UserCustomData: c.CustomData,
		Request:        newRequestData(c.Request),
		User:           User{c.User},
		Context:        Context{c.Identifier()},
		Client:         ClientData{"raygun4go", packageVersion, "https://github.com/MindscapeHQ/raygun4go"},
	}
}

// errorData is the struct holding all technical information on the error.
type ErrorData struct {
	Message    string     `json:"message"`    // the actual message the error produced
	StackTrace StackTrace `json:"stackTrace"` // the error's stack trace
}

// newErrorData fills returns a struct with all the information known about the
// error.
func newErrorData(err error, s StackTrace) ErrorData {
	return ErrorData{
		Message:    err.Error(),
		StackTrace: s,
	}
}

// currentStack returns the current stack. However, it omits the first 3 entries
// to avoid cluttering the trace with raygun4go-specific calls.
func currentStack() StackTrace {
	s := make(StackTrace, 0, 0)
	Current(&s)
	return s[3:]
}

// stackTraceElement is one element of the error's stack trace. It is filled by
// stack2struct.
type StackTraceElement struct {
	LineNumber  int    `json:"lineNumber"`
	PackageName string `json:"className"`
	FileName    string `json:"fileName"`
	MethodName  string `json:"methodName"`
}

// stackTrace is the stack the trace will be parsed into.
type StackTrace []StackTraceElement

// AddEntry is the method used by stack2struct to dump parsed elements.
func (s *StackTrace) AddEntry(lineNumber int, packageName, fileName, methodName string) {
	*s = append(*s, StackTraceElement{lineNumber, packageName, fileName, methodName})
}

// requestData holds all information on the request from the context
type RequestData struct {
	HostName    string            `json:"hostName"`
	URL         string            `json:"url"`
	HTTPMethod  string            `json:"httpMethod"`
	IPAddress   string            `json:"ipAddress"`
	QueryString map[string]string `json:"queryString"` // key-value-pairs from the URI parameters
	Form        map[string]string `json:"form"`        // key-value-pairs from a given form (POST)
	Headers     map[string]string `json:"headers"`     // key-value-pairs from the header
}

// newRequestData parses all information from the request in the context to a
// struct. The struct is empty if no request was set.
func newRequestData(r *http.Request) RequestData {
	if r == nil {
		return RequestData{}
	}

	r.ParseForm()

	return RequestData{
		HostName:    r.Host,
		URL:         r.URL.String(),
		HTTPMethod:  r.Method,
		IPAddress:   r.RemoteAddr,
		QueryString: arrayMapToStringMap(r.URL.Query()),
		Form:        arrayMapToStringMap(r.PostForm),
		Headers:     arrayMapToStringMap(r.Header),
	}
}

// clientData is the struct holding information on this client.
type ClientData struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	ClientURL string `json:"clientUrl"`
}

// user holds information on the affected user.
type User struct {
	Identifier string `json:"identifier"`
}

// context holds information on the program context.
type Context struct {
	Identifier string `json:"identifier"`
}

// UserCustomData is the interface that needs to be implemented by the custom
// data to be sent with the error. Being 'interface{}' suggests that it could
// be anything, but the data itself or contained data should respond to
// json.Marshal() for the data to be transmitted.
type UserCustomData interface{}
