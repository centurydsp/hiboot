# Hiboot - web/cli application framework 

<p align="center">
  <img src="https://github.com/hidevopsio/hiboot/blob/master/hiboot.png?raw=true" alt="hiboot">
</p>

<p align="center">
  <a href="https://travis-ci.org/hidevopsio/hiboot?branch=master">
    <img src="https://travis-ci.org/hidevopsio/hiboot.svg?branch=master" alt="Build Status"/>
  </a>
  <a href="https://codecov.io/gh/hidevopsio/hiboot">
    <img src="https://codecov.io/gh/hidevopsio/hiboot/branch/master/graph/badge.svg" />
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0">
      <img src="https://img.shields.io/badge/License-Apache%202.0-green.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/hidevopsio/hiboot">
      <img src="https://goreportcard.com/badge/github.com/hidevopsio/hiboot" />
  </a>
  <a href="https://godoc.org/github.com/hidevopsio/hiboot">
      <img src="https://godoc.org/github.com/golang/gddo?status.svg" />
  </a>
</p>

## About

Hiboot is a cloud native web and cli application framework written in Go.

Hiboot is not trying to reinvent everything, it integrates the popular libraries but make them simpler, easier to use. It borrowed some of the Spring features like dependency injection, aspect oriented programming, and auto configuration. You can integrate any other libraries easily by auto configuration with dependency injection support.

If you are a Java developer, you can start coding in Go without learning curve.

## Overview

* Web MVC (Model-View-Controller).
* Auto Configuration, pre-create instance with properties configs for dependency injection.
* Dependency injection with struct tag name **\`inject:""\`** or **Constructor** func.

## Features

* **Apps**
    * cli - command line application
    * web - web application

* **Starters**
    * actuator - health check
    * locale - locale starter
    * logging - customized logging settings
    * jwt - jwt starter
    * grpc - grpc application starter

* **Tags** 
    * inject - inject generic instance into object
    * default - inject default value into struct object 
    * value - inject string value or references / variables into struct string field
    * cmd - inject command into parent command for cli application
    * flag - inject flag / options into command object

* **Utils** 
    * cmap - concurrent map
    * copier - copy between struct
    * crypto - aes, base64, md5, and rsa encryption / decryption
    * gotest - go test util
    * idgen - twitter snowflake id generator
    * io - file io util
    * mapstruct - convert map to struct
    * replacer - replacing stuct field value with references or environment variables
    * sort - sort slice elements
    * str - string util enhancement util
    * validator - struct field validation 
       
and more features on the wey ...

## Getting started

This section will show you how to create and run a simplest hiboot application. Let’s get started!

### Getting started with Hiboot web application

#### Get the source code

```bash
go get -u github.com/hidevopsio/hiboot

cd $GOPATH/src/github.com/hidevopsio/hiboot/examples/web/helloworld/


```

#### Sample code
 
Below is the simplest web application in Go.


```go
// Line 1: main package
package main

// Line 2: import web starter from hiboot
import "github.com/hidevopsio/hiboot/pkg/app/web"

// Line 3-5: RESTful Controller, derived from web.Controller. The context mapping of this controller is '/' by default
type Controller struct {
	web.Controller
}

// Line 6-8: Get method, the context mapping of this method is '/' by default
// the Method name Get means that the http request method is GET
func (c *Controller) Get() string {
	// response data
	return "Hello world"
}

// Line 9-11: main function
func main() {
	// create new web application and run it
	web.NewApplication(&Controller{}).Run()
}
```

#### Run web application

```bash
dep ensure

go run main.go
```

#### Testing the API by curl

```bash
curl http://localhost:8080/
```

```
Hello, world
```

### Getting started with Hiboot cli application

Writing Hiboot cli application is as simple as web application, you can take the advantage of dependency injection introduced by Hiboot.

e.g. flag tag dependency injection

```go

// declare main package
package main

// import cli starter and fmt
import "github.com/hidevopsio/hiboot/pkg/app/cli"
import "fmt"

// define the command
type HelloCommand struct {
	// embedding cli.BaseCommand in each command
	cli.BaseCommand
	// inject (bind) flag to field 'To', so that it can be used on Run method, please note that the data type must be pointer
	To *string `flag:"name=to,shorthand=t,value=world,usage=e.g. --to=world or -t world"`
}

// Init constructor
func (c *HelloCommand) Init() {
	c.Use = "hello"
	c.Short = "hello command"
	c.Long = "run hello command for getting started"
}

// Run run the command
func (c *HelloCommand) Run(args []string) error {
	fmt.Printf("Hello, %v\n", *c.To)
	return nil
}

// main function
func main() {
	// create new cli application and run it
	cli.NewApplication(new(HelloCommand)).Run()
}

```

#### Run cli application

```bash
dep ensure

go run main.go
```

```bash
Hello, world
```

#### Build the cli application and run

```bash
go build
```

Let's get help

```bash
./hello --help
```

```bash
run hello command for getting started

Usage:
  hello [flags]

Flags:
  -h, --help        help for hello
  -t, --to string   e.g. --to=world or -t world (default "world")

```

Greeting to Hiboot

```bash
./hello --to Hiboot
```

```bash
Hello, Hiboot
```

### Dependency injection in Go

Dependency injection is a concept valid for any programming language. The general concept behind dependency injection is called Inversion of Control. According to this concept a struct should not configure its dependencies statically but should be configured from the outside.

Dependency Injection design pattern allows us to remove the hard-coded dependencies and make our application loosely coupled, extendable and maintainable.

A Go struct has a dependency on another struct, if it uses an instance of this struct. We call this a struct dependency. For example, a struct which accesses a user controller has a dependency on user service struct.

Ideally Go struct should be as independent as possible from other Go struct. This increases the possibility of reusing these struct and to be able to test them independently from other struct.

The following example shows a struct which has no hard dependencies.

```go

package main

// import cli starter and fmt
import (
	"fmt"
	"github.com/hidevopsio/hiboot/pkg/app"
	"github.com/hidevopsio/hiboot/pkg/app/cli"
)

// define the command
type rootCommand struct {
	// embedding cli.BaseCommand in each command
	cli.BaseCommand
	// inject (bind) flag to field 'To', so that it can be used on Run method, please note that the data type must be pointer
	To string `flag:"name=to,shorthand=t,value=world,usage=e.g. --to=world or -t world"`
}

func newRootCommand() *rootCommand {
	c := new(rootCommand)
	c.Use = "hello"
	c.Short = "hello command"
	c.Long = "run hello command for getting started"
	c.Example = `
hello -h : help
hello -t John : say hello to John
`
	c.PersistentFlags().StringVarP(&c.To, "to", "t", "world", "e.g. --to=world or -t world")
	return c
}

// Run run the command
func (c *rootCommand) Run(args []string) error {
	fmt.Printf("Hello, %v\n", c.To)
	return nil
}

// main function
func main() {
	// create new cli application and run it
	cli.NewApplication(newRootCommand).
		SetProperty(app.PropertyBannerDisabled, true).
		Run()
}

```

## Development Guide

### Git workflow
Below, we outline one of the more common Git workflows that core developers use. Other Git workflows are also valid.

### Fork the main repository

* Go to https://github.com/hidevopsio/hiboot
* Click the "Fork" button (at the top right)

### Clone your fork

The commands below require that you have $GOPATH set ($GOPATH docs). We highly recommend you put Istio's code into your GOPATH. Note: the commands below will not work if there is more than one directory in your $GOPATH.

```bash
export GITHUB_USER=your-github-username
mkdir -p $GOPATH/src/github.com/hidevopsio
cd $GOPATH/src/github.com/hidevopsio
git clone https://github.com/$GITHUB_USER/hiboot
cd hiboot
git remote add upstream 'https://github.com/hidevopsio/hiboot'
git config --global --add http.followRedirects 1
```

### Create a branch and make changes

```bash
git checkout -b my-feature
# Then make your code changes
```

### Keeping your fork in sync

```bash
git fetch upstream
git rebase upstream/master
```

Note: If you have write access to the main repositories (e.g. github.com/hidevopsio/hiboot), you should modify your Git configuration so that you can't accidentally push to upstream:
