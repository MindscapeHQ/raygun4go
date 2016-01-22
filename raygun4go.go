// Package raygun4go adds Raygun-based error handling to your golang code.
//
// It basically adds an error-handler that recovers from all panics that
// might occur and sends information about that error to Raygun. The amount
// of data being sent is configurable.
//
// Basic example:
//   raygun, err := raygun4go.New("appName", "apiKey")
//   if err != nil {
//     log.Println("Unable to create Raygun client:", err.Error())
//   }
//   defer raygun.HandleError()
//
// This will send the error message together with a stack trace to Raygun.
//
// However, raygun4go really starts to shine if used in a webserver context.
// By calling
//
//   raygun.Request(*http.Request)
//
// you can set a request to be analyzed in case of an error. If an error
// occurs, this will send the request details to Raygun, including
//
//   * hostname
//   * url
//   * http method
//   * ip adress
//   * url parameters
//   * POSTed form fields
//   * headers
//   * cookies
//
// giving you a lot more leverage on your errors than the plain error message
// could provide you with.
//
// Chainable configuration methods are available (see below) to set the
// affected version, user, tags or custom data.
package raygun4go

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/pborman/uuid"
)

// Client is the struct holding your Raygun configuration and context
// information that is needed if an error occurs.
type Client struct {
	appName     string              // the name of the app
	apiKey      string              // the api key for your raygun app
	context     *contextInformation // optional context information
	silent      bool                // if true, the error is printed instead of sent to Raygun
	logToStdOut bool
}

// contextInformation holds optional information on the context the error
// occured in.
type contextInformation struct {
	Version     string   // the version of the app
	DefaultTags []string // default tags that you would like to use to filter all the errors e.g. Production
	identifier  string   // a unique identifier for the running process, automatically set by New()
}

type ErrorEntry struct {
	request    *http.Request // the request associated to the error
	tags       []string      // tags that you would like to use to filter this error
	customData interface{}   // whatever you like Raygun to know about this error
	user       string        // the user that saw the error
	version    string        // the version of the app
	identifier string        // a unique identifier for the running process, automatically set by New()
	err        error
}

// raygunAPIEndpoint  holds the REST - JSON API Endpoint address
var raygunEndpoint = "https://api.raygun.io"

// Identifier returns the otherwise private identifier property from the
// Client's context. It is set by the New()-method and represents a unique
// identifier for your running program.
func (ci *contextInformation) Identifier() string {
	return ci.identifier
}

// New creates and returns a Client, needing an appName, appVersion and an apiKey. It also
// creates a unique identifier for you program.
func New(apiKey, appName, appVersion string, defaultTags []string) (c *Client, err error) {
	context := &contextInformation{identifier: uuid.New(), Version: appVersion, DefaultTags: defaultTags}
	if appName == "" || apiKey == "" {
		return nil, errors.New("appName and apiKey are required")
	}
	c = &Client{appName, apiKey, context, false, false}
	return c, nil
}

// Silent sets the silent-property on the Client. If true, errors will not be
// sent to Raygun but printed instead.
func (c *Client) Silent(s bool) *Client {
	c.silent = s
	return c
}

// LogToStdOut sets the logToStdOut-property on the Client.  If true, errors will
// be printed to std out as they are submitted to raygun.  This will also log
// any errors that occur when submiting to raygun to std out
func (c *Client) LogToStdOut(l bool) *Client {
	c.logToStdOut = l
	return c
}

func (c *Client) CreateErrorEntryFromMsg(msg string) *ErrorEntry {
	return c.CreateErrorEntry(errors.New(msg))
}

func (c *Client) CreateErrorEntry(err error) *ErrorEntry {
	entry := &ErrorEntry{
		err:        err,
		version:    c.context.Version,
		identifier: c.context.identifier,
	}
	entry.tags = append(entry.tags, c.context.DefaultTags...)
	return entry
}

// Request is a chainable option-setting method to add a request to the entry.
func (e *ErrorEntry) Request(r *http.Request) *ErrorEntry {
	e.request = r
	return e
}

// AppendTags is a chainable option-setting method to append tags to the entry. You
// can use tags to filter errors in Raygun.
func (e *ErrorEntry) AppendTags(tags []string) *ErrorEntry {
	e.tags = append(e.tags, tags...)
	return e
}

// CustomData is a chainable option-setting method to add arbitrary custom data
// to the entry. Note that the given type (or at least parts of it)
// must implement the Marshaler-interface for this to work.
func (e *ErrorEntry) CustomData(data interface{}) *ErrorEntry {
	e.customData = data
	return e
}

// User is a chainable option-setting method to add an affected Username to the
// entry.
func (e *ErrorEntry) User(u string) *ErrorEntry {
	e.user = u
	return e
}

// HandleError sets up the error handling code. It needs to be called with
//
//   defer c.HandleError()
//
// to handle all panics inside the calling function and all calls made from it.
// Be sure to call this in your main function or (if it is webserver) in your
// request handler as soon as possible.
func (c *Client) HandleError() error {
	e := recover()
	if e == nil {
		return nil
	}

	err, ok := e.(error)
	if !ok {
		err = errors.New(e.(string))
	}

	if c.logToStdOut {
		log.Println("Recovering from:", err.Error())
	}

	post := newPostData(c.CreateErrorEntry(err), currentStack())
	err = c.submit(post)

	if c.logToStdOut && err != nil {
		log.Println(err.Error())
	}
	return err
}

// SubmitError is a simple wrapper to manually post error entry (errors) to raygun
func (c *Client) SubmitError(entry *ErrorEntry) error {
	post := newPostData(entry, currentStack())
	return c.submit(post)
}

// submit takes care of actually sending the error to Raygun unless the silent
// option is set.
func (c *Client) submit(post postData) error {
	if c.silent {
		enc, _ := json.MarshalIndent(post, "", "\t")
		fmt.Println(string(enc))
		return nil
	}

	json, err := json.Marshal(post)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to convert to JSON (%s): %#v", err.Error(), post)
		return errors.New(errMsg)
	}

	r, err := http.NewRequest("POST", raygunEndpoint+"/entries", bytes.NewBuffer(json))
	if err != nil {
		errMsg := fmt.Sprintf("Unable to create request (%s)", err.Error())
		return errors.New(errMsg)
	}
	r.Header.Add("X-ApiKey", c.apiKey)
	httpClient := http.Client{}
	resp, err := httpClient.Do(r)

	defer resp.Body.Close()
	if resp.StatusCode == 202 {
		if c.logToStdOut {
			log.Println("Successfully sent message to Raygun")
		}
		return nil
	}

	errMsg := fmt.Sprintf("Unexpected answer from Raygun %d", resp.StatusCode)
	if err != nil {
		errMsg = fmt.Sprintf("%s: %s", errMsg, err.Error())
	}

	return errors.New(errMsg)
}
