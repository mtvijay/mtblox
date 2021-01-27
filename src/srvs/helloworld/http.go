package main

import (
	"context"
	"mime"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
        "strconv"
	"time"
	"reflect"

        "github.com/mtbox/mtlog"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"net/http/cookiejar"
)

const (
	MaxRecordLoggerLimit int = 100
)

const (
	ContentTypeProtoText      = "application/x-proto-text"
	ContentTypeProtoBinary    = "application/x-proto-binary"
	ContentTypeOctetStream    = "application/octet-stream"
	ContentTypeJSON           = "application/json"
	ContentTypeXML            = "application/xml"
	ContentTypeHTML           = "text/html"
	ContentTypeText           = "text/plain"
	ContentTypeCspReport      = "application/csp-report"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	ContentTypeOcsp           = "application/ocsp-report"
)

type Mcontext struct {
	ctx    context.Context
	w      http.ResponseWriter
	r      *http.Request
	params map[string]string
}



//MachesContentType validates the content type against the expected one
func MatchesContentType(contentType, expectedType string) bool {
	mimeType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mtlog.Errorf("Error parsing media type: %s error %v", contentType, err)
	}

	return err == nil && mimeType == expectedType
}

// SetCommonHeader sets a common set of headers in HTTP response message
func (m *Mcontext) SetCommonHeader() {
	if m.w != nil {
		setCommonHeader(m.r, m.w)
		return
	}
}

// WriteCode writes HTTP status code in HTTP response message
func (m *Mcontext) WriteCode(code int) {
	m.SetCommonHeader()
	if m.w != nil {
		m.w.WriteHeader(code)
	}
}

func (m *Mcontext) ClearParams() error {
	if m.r.MultipartForm != nil {
		return m.r.MultipartForm.RemoveAll()
	}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", ContentTypeJSON)

	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func (m *Mcontext) WriteJSON(code int, v interface{}) error {
	m.SetCommonHeader()
	return writeJSON(m.w, code, v)
}

// CheckForJSON makes sure that the request's Content-Type is application/json
func checkForJSON(r *http.Request) error {
	ct := r.Header.Get("Content-Type")

	// No Content-Type header is ok as long as there's no body
	if ct == "" {
		if r.Body == nil || r.ContentLength == 0 {
			return nil
		}
	}

	if MatchesContentType(ct, ContentTypeJSON) {
		return nil
	}
	return fmt.Errorf("Content-Type specified (%s) mst be %s", ct, ContentTypeJSON)
}

func readJSON(r *http.Request, w http.ResponseWriter, v interface{}, limit int) error {
	if err := checkForJSON(r); err != nil {
		return err
	}

	if limit != 0 {
		contentLength := r.Header.Get("Content-Length")
		if contentLength == "" {
			http.Error(w, http.StatusText(http.StatusLengthRequired), http.StatusLengthRequired)
			return fmt.Errorf(http.StatusText(http.StatusLengthRequired))
		}

		length, err := strconv.Atoi(contentLength)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest),
				http.StatusBadRequest)
			return fmt.Errorf(http.StatusText(http.StatusBadRequest))
		}

		if length > limit {
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return fmt.Errorf(http.StatusText(http.StatusRequestEntityTooLarge))
		}
	}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(v)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	return err
}

func (m *Mcontext) ReadJSON(v interface{}) error {
	return readJSON(m.r, m.w, v, 0)
}
// Requests...
type HutilsRequest struct {
	url          string        // Raw URL string
	Method       string        // HTTP method to use
	Params       map[string]string // Params list
	Payload      interface{}
	Insecure     bool // Uses http (not https) when true
	SkipResponse bool // Let the end user process the response

	// response from server is unmarshaled into Result.
	Result interface{}

	// Optional
	Userinfo *url.Userinfo
	Header   *http.Header

	// The following fields are populated by Send().
	timestamps time.Time // Time when HTTP request was sent
	timestampr time.Time // Time when HTTP request was sent

	status   int            // HTTP status for executed request
	response *http.Response // HutilsResponse object from http package
	body     []byte         // Body of server's response (JSON or otherwise)
}

// Per session information
type HutilsConnect struct {
	sync.Mutex

	name string

	// Flight recording, store last few requests with Result
	Record bool
	Logs   []HutilsRequest

	// Optional
	Userinfo *url.Userinfo
	Header   *http.Header

	certPool    *x509.CertPool
	certCount   int
	sslClient   *http.Client
	httpClient  *http.Client

	// minimum client version in int format
	minClientVersion int

	r *mux.Router
}

type HttpApiFunc struct {
	Fn  func(*Mcontext) (int, interface{})
	Enc string
}

type HttpApiHandleArray map[string]map[string]HttpApiFunc

type HttpRouteLayout struct {
	Handlers HttpApiHandleArray
	Group    string
	Prefix   string
}


// This function replaces the cert pool with the given list of certs.
// The previous certs are lost
func (c *HutilsConnect) replaceCertPool(certs *[]*x509.Certificate) {
	if certs == nil || len(*certs) == 0 {
		return
	}

	c.certPool = x509.NewCertPool()

	for _, v := range *certs {
		c.certPool.AddCert(v)
		c.certCount++
	}
}

// fixme: This is when wrap all the client in a standard library we can figure out we are running a lower version
// library. We should also log the ip address of the incoming caller here

// FIX-ME : Separate out HTTP handler for login, device requests and UI/CLI requests
// Avoid using localRoute as that can change
func makeHttpHandler(h *HutilsConnect, localMethod string, localRoute string, handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx              context.Context
			cancel           context.CancelFunc
		)
		// FIXME, check for r.FormValue is not null
		timeout, err := time.ParseDuration(r.FormValue("timeout"))
		if err == nil {
			// The request has a timeout, so create a context that is
			// canceled automatically when the timeout expires.
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		defer cancel() // Cancel ctx as soon as handleSearch returns.

		now := time.Now()
/*
		ctx = context.WithValue(ctx, ClientIPKey, r.Header.Get("X-REAL-IP"))
		ctx = context.WithValue(ctx, ServerHostKey, r.Header.Get("X-HOST"))
		ctx = context.WithValue(ctx, RequestIDKey, r.Header.Get("X-Request-Id"))
		ctx = context.WithValue(ctx, RequestMethodKey, localMethod)
*/
		mctx := &Mcontext{ctx: ctx, w: w, r: r}
		code, intf := handlerFunc.Fn(mctx)
		traceHttpReqStatus(r, code, now, "")
		if intf != nil {
			// Always work with pointers. If a struct is returned by the handler
			// func, it will never resolve to proto.Message. Convert to a pointer
			// before the encoding.
			if reflect.ValueOf(intf).Kind() == reflect.Struct {
				vp := reflect.New(reflect.TypeOf(intf))
				vp.Elem().Set(reflect.ValueOf(intf))
				intf = vp.Interface()
			}

			switch eIntf := intf.(type) {
			case error:
				mctx.SetCommonHeader()
				http.Error(w, eIntf.Error(), code)
			case string:
				mctx.SetCommonHeader()
				http.Error(w, eIntf, code)
			default:
				mctx.WriteJSON(code, eIntf)
			}
		} else {
			mctx.SetCommonHeader()
			mctx.WriteCode(code)
		}
		_ = mctx.ClearParams()
	}
}

func (h *HutilsConnect) AppendRouter(m HttpRouteLayout) error {
	subRouter := h.r.PathPrefix(m.Prefix).Subrouter()
	for method, routes := range m.Handlers {
		for route, fct := range routes {
			localRoute := m.Prefix + route
			mtlog.Tracef("Registering %s, %s", method, localRoute)

			// build the handler function
			f := makeHttpHandler(h, method, localRoute, fct)

			// add the new route
			subRouter.Path(route).Methods(method).HandlerFunc(f)
		}
	}

	return nil
}

func traceHttpReqStatus(r *http.Request, code int, opStartTime time.Time, errStr string) {
	mtlog.Tracef("Request ID %v , proxy Request ID %v \"%s %s %s\" from IP %v \"%v\" returned %v (%v) after %v",
		r.Header.Get("X-Request-Id"), r.Header.Get("X-Proxy-Request-Id"), r.Method,
		r.URL.String(), r.Proto, r.Header.Get("X-REAL-IP"),
		r.UserAgent(), code, errStr, time.Now().Sub(opStartTime))
}

func setHeader(w http.ResponseWriter, key string, value string) {
	w.Header().Set(key, value)
	return
}


func setCommonHeader(r *http.Request, w http.ResponseWriter) {
	setHeader(w, "Content-Security-Policy", "default-src 'self'; base-uri 'self'; connect-src 'self'  https://maps.googleapis.com https://maps.gstatic.com; child-src 'self'; font-src 'self'; form-action 'self'; frame-ancestors 'none'; frame-src 'self'; img-src * data: maps.gstatic.com *.googleapis.com *.ggpht; script-src 'self' https://maps.googleapis.com https://maps.gstatic.com; style-src 'self'")
	setHeader(w, "Referrer-Policy", "strict-origin-when-cross-origin")
	setHeader(w, "X-Content-Type-Options", "nosniff")
	setHeader(w, "X-Frame-Options", "DENY")
	if token := csrf.Token(r); token != "" {
		setHeader(w, "X-CSRF-Token", token)
	}
}

// HTTP 401 "Unauthorized" handler function
func zedUnauthorizedHandler(w http.ResponseWriter, r *http.Request, errStr string) {
	now := time.Now()
	response := ZsrvUnauthorized()
	respJSON, _ := json.Marshal(response)
	setCommonHeader(r, w)
	setHeader(w, "Content-Type", "application/json")
	//The server generating 401 response MUST send a WWW-Authenticate header
	//field containing at least one challenge applicable to target resource.
	setHeader(w, "WWW-Authenticate", "Bearer")
	w.WriteHeader(int(response.HttpStatusCode))
	w.Write([]byte(respJSON))
	traceHttpReqStatus(r, int(response.HttpStatusCode), now, errStr)
}

// HTTP 403 "Forbidden" handler function
func zedForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	var errStr string
	now := time.Now()
	response := ZsrvForbidden()
	if err := csrf.FailureReason(r); err != nil {
		errStr = err.Error()
	}
	respJSON, _ := json.Marshal(response)
	setCommonHeader(r, w)
	setHeader(w, "Content-Type", "application/json")
	w.WriteHeader(int(response.HttpStatusCode))
	w.Write([]byte(respJSON))
	traceHttpReqStatus(r, int(response.HttpStatusCode), now, errStr)
}

// HTTP 404 "Not Found" handler function
func zedNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	response := ZsrvObjNotFound()
	respJSON, _ := json.Marshal(response)
	setCommonHeader(r, w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.HttpStatusCode))
	w.Write([]byte(respJSON))
	traceHttpReqStatus(r, int(response.HttpStatusCode), now, response.HttpStatusMsg)
}

// HTTP 405 "Method Not Allowed" handler function
func zedMethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	response := ZsrvMethodNotAllowed()
	respJSON, _ := json.Marshal(response)
	/* To-Do: Find how to be compatible with RFC 7231
	The origin server MUST generate an Allow header field in a 405 response
	containing a list of the target resource's currently supported methods.
	*/
	setCommonHeader(r, w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.HttpStatusCode))
	w.Write([]byte(respJSON))
	traceHttpReqStatus(r, int(response.HttpStatusCode), now, response.HttpStatusMsg)
}


func (h *HutilsConnect) ListenAndServe(addr string, csrfProtect bool, csrfPath string) error {
	httpSrv := http.Server{}
	if h.r == nil {
		return fmt.Errorf("no router context")
	}

	originsOk := handlers.AllowedOrigins([]string{""})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	if csrfProtect == true {
		csrfKey := uuid.New()
		csrfMiddleware := csrf.Protect(
			csrfKey[:],
			csrf.Secure(false),
			csrf.HttpOnly(false),
			csrf.ErrorHandler(http.HandlerFunc(zedForbiddenHandler)),
			csrf.Path(csrfPath),
		)
		optionsOk := handlers.IgnoreOptions()
		httpSrv = http.Server{Addr: addr, Handler: csrfMiddleware(handlers.CORS(originsOk, methodsOk, optionsOk)(h.r))}
	} else {
		httpSrv = http.Server{Addr: addr, Handler: handlers.CORS(originsOk, methodsOk)(h.r)}
	}
	httpSrv.SetKeepAlivesEnabled(true)

	return httpSrv.ListenAndServe()
}

func NewConnect(name string, record bool, certs *[]*x509.Certificate) (*HutilsConnect, error) {
	vCtx := &HutilsConnect{name: name, Record: record}

	// FIXME: DisableCompresion should per request setting
	vCtx.replaceCertPool(certs)
	t := &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
		DisableCompression:    true,
	}
	vCtx.sslClient = &http.Client{Transport: t}

	var err error
	vCtx.sslClient.Jar, err = cookiejar.New(nil)
	if err == nil {
		vCtx.httpClient.Jar, err = cookiejar.New(nil)
	}

	//Recorder
	if record {
		vCtx.Logs = make([]HutilsRequest, MaxRecordLoggerLimit)
	}

	vCtx.r = mux.NewRouter()
	// HTTP error code 404 - NotFound handler
	vCtx.r.NotFoundHandler = http.HandlerFunc(zedNotFoundHandler)
	// HTTP error code 405 - MethodNotAllowed handler
	vCtx.r.MethodNotAllowedHandler = http.HandlerFunc(zedMethodNotAllowedHandler)

	return vCtx, err
}
