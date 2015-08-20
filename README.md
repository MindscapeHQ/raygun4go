# raygun4goGAE
[![Build Status](https://travis-ci.org/miko-code/raygun4goGAE.svg?branch=master)](https://travis-ci.org/miko-code/raygun4goGAE)
[![Build status](https://ci.appveyor.com/api/projects/status/9pqk769jaxfxp0bb/branch/master?svg=true)](https://ci.appveyor.com/project/kaeuferportal-oss/raygun4go/branch/master)
[![Coverage](http://gocover.io/_badge/github.com/miko-code/raygun4goGAE)](http://gocover.io/github.com/miko-code/raygun4goGAE)
[![GoDoc](https://godoc.org/github.com/miko-code/raygun4goGAE?status.svg)](http://godoc.org/github.com/miko-code/raygun4goGAE)
[![GoReportcard](http://goreportcard.com/badge/miko-code/raygun4goGAE)](http://goreportcard.com/report/miko-code/raygun4goGAE)


raygun4go adds Raygun-based error handling to your golang code. It catches all
occuring errors, extracts as much information as possible and sends the error
to Raygun via their REST-API.

## Getting Started

### Installation
```
  $ go get github.com/miko-code/raygun4goGAE
```

### Basic Usage

Include the package and then defer the HandleError-method as soon as possible
in a context as global as possible. In webservers, this will probably be your
request handling method, in all other programs it should be your main-method.
Having found the right spot, just add the following example code:


```
//generate new appengine context
ctx := appengine.NewContext(r)

raygun, err := raygun4go.New("appName", "apiKey")
if err != nil {
  log.Println("Unable to create Raygun client:", err.Error())
}
raygun.Silent(true)
defer raygun.HandleError(ctx)

raygun.CreateError("this is error", ctx)


```

where ``appName`` is the name of your app and ``apiKey`` is your
Raygun-API-key. If your program runs into a panic now (which you can easily
test by adding ``panic("foo")`` after the call to ``defer``), the handler will
print the resulting error message. If you remove the line
```
raygun.Silent(true)
```
the error will be sent to Raygun using your API-key.

### Options

The client returned by ``New`` has several chainable option-setting methods

Method                    | Description
--------------------------|------------------------------------------------------------
`Silent(bool)`            | If set to true, this prevents the handler from sending the error to Raygun, printing it instead.
`Request(*http.Request)`  | Adds the responsible http.Request to the error.
`Version(string)`         | If your program has a version, you can add it here.
`Tags([]string)`          | Adds the given tags to the error. These can be used for filtering later.
`CustomData(interface{})` | Add arbitrary custom data to you error. Will only reach Raygun if it works with `json.Marshal()`.
`User(string)`            | Add the name of the affected user to the error.

## Bugs and feature requests

Have a bug or a feature request? Please first check the list of
[issues](https://github.com/miko-code/raygun4goGAE/issues).

If your problem or idea is not addressed yet, [please open a new
issue](https://github.com/miko-code/raygun4goGAE/issues/new).
