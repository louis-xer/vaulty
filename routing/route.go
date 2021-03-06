package routing

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
	"github.com/vaulty/vaulty/transformer"
)

type RouteParams struct {
	Name                    string
	Method                  string
	URL                     string
	Upstream                string
	RequestTransformations  []transformer.Transformer
	ResponseTransformations []transformer.Transformer
}

type Route struct {
	Name        string
	UpstreamURL *url.URL
	IsInbound   bool

	method                  string
	rawURL                  string
	url                     *url.URL
	requestTransformations  []transformer.Transformer
	responseTransformations []transformer.Transformer
	g                       glob.Glob
}

func NewRoute(params *RouteParams) (*Route, error) {
	var err error
	route := &Route{
		Name:                    params.Name,
		method:                  params.Method,
		rawURL:                  params.URL,
		requestTransformations:  params.RequestTransformations,
		responseTransformations: params.ResponseTransformations,
	}

	route.url, err = url.Parse(params.URL)
	if err != nil {
		return nil, err
	}

	route.IsInbound = !route.url.IsAbs()

	if route.IsInbound && params.Upstream == "" {
		return nil, fmt.Errorf("Missed Upstream for inbound route %s", params.Name)
	}

	route.UpstreamURL, err = url.Parse(params.Upstream)
	if err != nil {
		return nil, err
	}

	route.g, err = glob.Compile(route.rawURL)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (r *Route) Match(req *http.Request) bool {
	var matchingURL *url.URL

	// no need to do any checking for inbound request and outbound route
	if req.URL.Host == "inbound" && !r.IsInbound {
		return false
	}

	if req.URL.Host == "inbound" {
		matchingURL = &url.URL{}
		matchingURL.Path = req.URL.Path
	} else {
		matchingURL = &url.URL{}
		// for CONNECT target URI is not absolute url
		// so goproxy builds URL by using authority-form (HOST:PORT)
		// as request.Host. For https/443 we will remove port from
		// matchingURL.Host. If other port is specified, then it should
		// be used in route.url as well
		if req.URL.Port() == "443" && req.URL.Scheme == "https" {
			matchingURL.Host = req.URL.Hostname()
		} else {
			matchingURL.Host = req.URL.Host
		}

		matchingURL.Scheme = req.URL.Scheme
		matchingURL.Path = req.URL.Path
	}

	if matchingURL.Path == "" {
		matchingURL.Path = "/"
	}

	// check if route URL matches request URL
	// here we use filepath.Match which seems to be pretty good
	// for our goal.
	log.Debugf("Match route path %s against URL %s", r.rawURL, matchingURL.String())
	urlMatch := r.g.Match(matchingURL.String())

	return urlMatch && (r.method == "*" || req.Method == r.method)
}

func (r *Route) TransformRequest(req *http.Request) (*http.Request, error) {
	var err error

	for _, tr := range r.requestTransformations {
		req, err = tr.TransformRequest(req)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (r *Route) TransformResponse(res *http.Response) (*http.Response, error) {
	var err error

	for _, tr := range r.responseTransformations {
		res, err = tr.TransformResponse(res)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
