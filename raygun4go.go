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
	appName     string             // the name of the app
	apiKey      string             // the api key for your raygun app
	context     contextInformation // optional context information
	silent      bool               // if true, the error is printed instead of sent to Raygun
	logToStdOut bool
}

// contextInformation holds optional information on the context the error
// occured in.
type contextInformation struct {
	Request    *http.Request // the request associated to the error
	Version    string        // the version of the package
	Tags       []string      // tags that you would like to use to filter this error
	CustomData interface{}   // whatever you like Raygun to know about this error
	User       string        // the user that saw the error
	identifier string        // a unique identifier for the running process, automatically set by New()
}

// raygunAPIEndpoint  holds the REST - JSON API Endpoint address
var raygunEndpoint = "https://api.raygun.io"

// Identifier returns the otherwise private identifier property from the
// Client's context. It is set by the New()-method and represents a unique
// identifier for your running program.
func (ci *contextInformation) Identifier() string {
	return ci.identifier
}

// New creates and returns a Client, needing an appName and an apiKey. It also
// creates a unique identifier for you program.
func New(appName, apiKey string) (c *Client, err error) {
	context := contextInformation{identifier: uuid.New()}
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

// Request is a chainable option-setting method to add a request to the context.
func (c *Client) Request(r *http.Request) *Client {
	c.context.Request = r
	return c
}

// Version is a chainable option-setting method to add a version to the context.
func (c *Client) Version(v string) *Client {
	c.context.Version = v
	return c
}

// Tags is a chainable option-setting method to add tags to the context. You
// can use tags to filter errors in Raygun.
func (c *Client) Tags(tags []string) *Client {
	c.context.Tags = tags
	return c
}

// CustomData is a chainable option-setting method to add arbitrary custom data
// to the context. Note that the given type (or at least parts of it)
// must implement the Marshaler-interface for this to work.
func (c *Client) CustomData(data interface{}) *Client {
	c.context.CustomData = data
	return c
}

// User is a chainable option-setting method to add an affected Username to the
// context.
func (c *Client) User(u string) *Client {
	c.context.User = u
	return c
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

	post := c.createPost(err, currentStack())
	err = c.submit(post)

	if c.logToStdOut && err != nil {
		log.Println(err.Error())
	}
	return err
}

// createPost creates the data structure that will be sent to Raygun.
func (c *Client) createPost(err error, stack stackTrace) postData {
	return newPostData(c.context, err, stack)
}

// CreateError is a simple wrapper to manually post messages (errors) to raygun
func (c *Client) CreateError(message string) error {
	err := errors.New(message)
	post := c.createPost(err, currentStack())
	return c.submit(post)
}

//Post holds postData (including a stack). This allows for asynchronous reporting to raygun (without losing the stack trace)
type Post struct {
	postData postData
}

//CreatePost is a wrapper to manually post errors to raygun. It also allows to specify the index for stack trace truncation.
func (c *Client) CreatePost(err error, stackTruncateAt int) Post {
	post := c.createPost(err, currentStackAt(stackTruncateAt))
	return Post{postData: post}
}

//SubmitPost is a wrapper to manually post errors to raygun asynchronously (having previously captured a stack trace on a separate goroutine, using c.CreatePost())
func (c *Client) SubmitPost(post Post) error {
	return c.submit(post.postData)
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
