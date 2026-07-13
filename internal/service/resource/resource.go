package resource

// first return type needs to be a json parsable value or struct. as always, if error is returned ignore the first return value. only return unknown or I/O and handle other errors in the resource to keep creating some overview data.
type ExploreResource func(projectRootPath string, verbose bool) (any, error)
