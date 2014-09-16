package grunway

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amattn/deeperror"
)

// convenience struct holding all the stuff you usually want to know about an endpoint
type Endpoint struct {
	VersionStr string
	EntityName string
	PrimaryKey int64
	Action     string
	Extras     []string

	// internal only
	version        VersionUint
	versionConvErr error
}

// return a typed number, not a string
// cache value so we only do this once.
func (e *Endpoint) Version() VersionUint {
	if e.versionConvErr == nil && e.version == 0 {
		var v64 uint64
		v64, e.versionConvErr = strconv.ParseUint(e.VersionStr, 10, VERSION_BIT_DEPTH)
		e.version = VersionUint(v64)
	}
	return e.version
}

// exported for other packages to be able to unit test.
func ParsePathForTesting(urlPtr *url.URL, prefix string) (endpoint Endpoint, err error) {
	endpoint, clientErr, serverErr := parsePath(urlPtr, prefix)

	if serverErr != nil {
		err = serverErr
	}
	if clientErr != nil {
		err = clientErr
	}

	return endpoint, err
}

func parsePath(urlPtr *url.URL, prefix string) (endpoint Endpoint, clientErr, serverErr *deeperror.DeepError) {
	urlPath := strings.Trim(urlPtr.Path, "/")
	prefix = strings.TrimLeft(prefix, "/")

	if strings.HasPrefix(urlPath, prefix) == false {
		// is this an error?
		return Endpoint{}, deeperror.NewHTTPError(3475081071, "Invalid Prefix", nil, http.StatusNotFound), nil
	}
	urlPath = urlPath[len(prefix):]
	urlPath = strings.Trim(urlPath, "/")

	pathComponents := strings.Split(urlPath, "/")
	pathComponentsLen := len(pathComponents)

	// basic validation: should have at least a version and an entity
	if pathComponentsLen < 2 {
		return Endpoint{}, deeperror.NewHTTPError(3475081072, "Cannot parse endpoint path, insufficent number of path components", nil, http.StatusNotFound), nil
	}

	// parse version
	if pathComponentsLen >= 1 {
		s := strings.TrimLeft(pathComponents[0], "vV")
		s = strings.TrimLeft(s, "0")
		endpoint.VersionStr = s
	}

	// parse entity
	if pathComponentsLen >= 2 {
		endpoint.EntityName = pathComponents[1]
	}

	// parse pk and extra
	if pathComponentsLen >= 3 {
		// parse pk
		pkeyOrActionString := pathComponents[2]
		pkey, err := strconv.ParseInt(pkeyOrActionString, 10, 64)
		if err == nil {
			endpoint.PrimaryKey = pkey
			if pathComponentsLen >= 4 {
				endpoint.Action = pathComponents[3]
			}
		} else {
			//it's probably an action
			endpoint.Action = pkeyOrActionString
		}

		// parse extras
		endpoint.Extras = pathComponents[2:]
	}

	return
}
