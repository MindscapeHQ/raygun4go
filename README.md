# Raygun4Go
[![Coverage](http://gocover.io/_badge/github.com/MindscapeHQ/raygun4go)](http://gocover.io/github.com/MindscapeHQ/raygun4go)
[![GoDoc](https://godoc.org/github.com/MindscapeHQ/raygun4go?status.svg)](http://godoc.org/github.com/MindscapeHQ/raygun4go)
[![GoReportcard](http://goreportcard.com/badge/MindscapeHQ/raygun4go)](http://goreportcard.com/report/MindscapeHQ/raygun4go)

Raygun4Go adds Raygun-based error handling to your Golang code. It catches all occuring errors, extracts as much information as possible, and sends the error to Raygun via our REST-API.

## Getting Started

### Installation
```bash
$ go get github.com/MindscapeHQ/raygun4go
```

### Basic Usage

Include the package and then defer the `HandleError`-method as soon as possible, in a context as global as possible.
In webservers, this will probably be your request handling method. In all other programs, it should be your main-method.
It will automatically capture any _panic_ and report it.

```go
raygun, err := raygun4go.New("appName", "apiKey")
if err != nil {
  log.Println("Unable to create Raygun client:", err.Error())
}
defer raygun.HandleError()
```

Where ``appName`` is the name of your app and ``apiKey`` is your Raygun API key.
If your program runs into a panic now (which you can easily test by adding ``panic("foo")`` after the call to ``defer``), the handler will send the error to Raygun.

#### Manually sending errors

To send errors manually, you can use `CreateError(message string)`, `SendError(error error)`, or `CreateErrorWithStackTrace(message string, st StackTrace)`.
- `CreateError` creates an error with the given message and immediately reports it with the current execution stack trace.
- `SendError` immediately reports the error with its stack trace if it's of type `"github.com/go-errors/errors".Error`. Otherwise, it uses the current execution stack trace.
- `CreateErrorWithStackTrace` allows you to manually send an error with a custom stack trace.

---

Example of `CreateError`:
```go
if err := raygun.CreateError("something bad happened"); err != nil {
    log.Printf("failed to report error to Raygun: %v\n", err)
}
```

Example of `SendError`:
```go
err := something.Do()
if err := raygun.SendError(err); err != nil {
    log.Printf("failed to report error to Raygun: %v\n", err)
}
```

Example of `CreateErrorWithStackTrace`:
```go
st := make(raygun4go.StackTrace, 0)
st.AddEntry(42, "main", "example.go", "exampleFunc")
if err := raygun.CreateErrorWithStackTrace("something bad happened", st); err != nil {
    log.Printf("failed to report error to Raygun: %v\n", err)
}
```

---

### Options

The client returned by ``New`` has several chainable option-setting methods:

Method                    | Description
--------------------------|------------------------------------------------------------
`Silent(bool)`            | If set to `true`, this prevents the handler from sending the error to Raygun, printing it instead.
`Request(*http.Request)`  | Adds the responsible `http.Request` to the error.
`Version(string)`         | If your program has a version, you can add it here.
`Tags([]string)`          | Adds the given tags to the error. These can be used for filtering later.
`CustomData(interface{})` | Adds arbitrary custom data to you error. Will only reach Raygun if it works with `json.Marshal()`.
`User(string)`            | Adds the name of the affected user to the error.

### Custom grouping

By default, the Raygun service will group errors together based on stack trace content.
If you have any cases where you want to control the error grouping yourself, then you can provide a custom-grouping-key callback function.
Below is a simple example that returns a hard-coded grouping key, which would cause all errors to be grouped together:
```go
raygun.CustomGroupingKeyFunction(func(error, raygun4go.PostData)string{return "customGroupingKey"})
```

The callback takes the original error, and the Raygun `PostData` payload structure that is about to be serialized and sent to Raygun.
In your callback, you can check these values to help build your own grouping key logic based on different cases that you want to control.
For any error you don't want to group yourself, return an empty string - Raygun will then use the default grouping.

## Bugs and feature requests

Have a bug or a feature request? Please first check the list of [issues](https://github.com/MindscapeHQ/raygun4go/issues).

If your problem or idea is not addressed yet, [please open a new issue](https://github.com/MindscapeHQ/raygun4go/issues/new).
