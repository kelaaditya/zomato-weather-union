package utilities

import "net/url"

// generic function to check if a pointer is nil.
// if the pointer is not nil, dereference it to get value.
func DereferenceOrNil[T comparable](p *T) any {
	if p != nil {
		return *p
	}
	return nil
}

// function to build the URL string with parameters
func BuildURLString(
	baseURL string,
	customPath string,
	queryParameters map[string]string,
) (string, error) {
	// join base URL with specific path
	specificURLString, err := url.JoinPath(baseURL, customPath)
	if err != nil {
		return "", err
	}

	// parse to URL type
	specificURL, err := url.Parse(specificURLString)
	if err != nil {
		return "", err
	}

	// add query parameters to the URL object
	q := specificURL.Query()
	// iterate over query parameter map
	for k, v := range queryParameters {
		q.Set(k, v)
	}
	// build raw query
	specificURL.RawQuery = q.Encode()

	// get string of the final built API URL
	var finalURLString string = specificURL.String()

	return finalURLString, nil
}
