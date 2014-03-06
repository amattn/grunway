grunway
=======

Designed for quickly making REST APIs, w/ easy config and just enough magic to keep you productive, written in Go


Installation
------------

Same as every other go package:

	go get github.com/amattn/grunway


Usage
-----

Grunway is built around 2 central concepts:

1. Defining high-level routes
2. Quickly implementing handlers

Under the hood, routes map to endpoints which call handlers.
refelection is used during startup to build an internal routing map, but not during routing of requests.

### Defining High-level Routes

Looks like this:

	routerPtr := grunway.NewRouter()
	routerPtr.BasePath = "/api/"
	routerPtr.RegisterEntity("book", &BookController{})
	routerPtr.RegisterEntity("author", &AuthorController{})
	return routerPtr

You then start the server like so:

	host := ":8090"
	err := grunway.Start(routerPtr, host)
	if err != nil {
		log.Fatalln(err)
	}

### Quickly implementing handlers

Here is where we use a bit of reflection.  Instead of defining routes and hooking up controllers, we _just_ immplement handlers.

Like this:

	func (bookCon *BookController) GetHandlerV1(ctx *grunway.Context) (*grunway.RouteError, grunway.PayloadsMap, grunway.CustomRouteResponse) {
	   //...
	}
	func (bookCon *BookController) GetHandlerV2(ctx *grunway.Context) (*grunway.RouteError, grunway.PayloadsMap, grunway.CustomRouteResponse) {
		//...
	}
	func (bookCon *BookController) GetHandlerV1Popular(ctx *grunway.Context) (*grunway.RouteError, grunway.PayloadsMap, grunway.CustomRouteResponse) {
	    //...
	}

The basic function signature is a `RouteHandler`:

	type RouteHandler func(*Context) (*RouteError, PayloadsMap, CustomRouteResponse)

99% of the time you either return a PayloadsMap or a RouteError.  If you need special control of the response, a CustomRouteResponse is a special handler with more access to the output stream.


The format of the Handlers works like this:

	<Method>HandlerV<version><Action>

Which corresponts to http endpoints like this:

    <Method> http://host/<prefix>/v<version>/<Entity>/<optionalID>/<Action>

### Auth

Auth is a bit special that it has its own dedicated handler prefix:

	func (con *WidgetController) AuthGetHandlerV1(ctx *grunway.Context) (*grunway.RouteError, grunway.PayloadsMap, grunway.CustomRouteResponse) {
	   //...
	}

There is a dedicated package for handling auth at http://github.com/amattn/grwacct

