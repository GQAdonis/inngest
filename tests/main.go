package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/inngest/inngest/inngest"
	"github.com/inngest/inngest/pkg/sdk"
	"github.com/stretchr/testify/require"
)

const (
	ENV_SDK_URL     = "SDK_URL"   // eg. http://127.0.0.1:3000/api/inngest
	ENV_API_URL     = "API_URL"   // eg http://127.0.0.1:8288 or https://api.inngest.com
	ENV_EVENT_URL   = "EVENT_URL" // eg http://127.0.0.1:8288 or https://inn.gs
	ENV_SIGNING_KEY = "INNGEST_SIGNING_KEY"
	ENV_EVENT_KEY   = "INNGEST_EVENT_KEY"
	ENV_PROXY_URL   = "PROXY_URL"
)

var (
	sdkURL               url.URL
	apiURL               url.URL
	eventURL             url.URL
	proxyURL             string
	signingKey, eventKey string

	buffer = 5 * time.Second
)

func init() {
	sdkURL = parseEnvURL(ENV_SDK_URL)
	apiURL = parseEnvURL(ENV_API_URL)
	// If the ENV_EVENT_URL isn't specified default to the API url, assuming that this
	// is the dev server and runs both APIs.
	if os.Getenv(ENV_EVENT_URL) == "" {
		os.Setenv(ENV_EVENT_URL, apiURL.String())
	}
	eventURL = parseEnvURL(ENV_EVENT_URL)
	signingKey = os.Getenv(ENV_SIGNING_KEY)
	eventKey = os.Getenv(ENV_EVENT_KEY)
	if eventKey == "" {
		eventKey = "eventkey"
	}

	proxyURL = os.Getenv(ENV_PROXY_URL)
	if proxyURL == "" {
		proxyURL = "http://127.0.0.1:42018"
	}
}

func parseEnvURL(env string) url.URL {
	u, err := url.Parse(os.Getenv(env))
	if err != nil {
		panic(err.Error())
	}
	if u.Host == "" {
		panic(fmt.Sprintf("No %s env variable supplied", env))
	}
	return *u
}

func run(t *testing.T, test *Test) {
	t.Helper()

	test.requests = make(chan http.Request)
	test.responses = make(chan http.Response)

	// Ensure that the desired function exists within the SDK.
	rr, err := introspect(test)
	require.NoError(t, err, "Introspection error")

	// Start a new test server which will intercept all requests between the executor and the SDK.
	//
	// This allows us to assert that the messages passed are as we expect, including:
	//
	// - Responses from the SDK
	// - State injected via the executor.
	//
	// We can also randomly inject faults by disregarding the SDK's response and throwing a 500.
	mux := http.NewServeMux()
	mux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		byt, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		fmt.Printf(" ==> Received executor request:\n\t%s\n", string(byt))

		// Recreate reader to re-read in assertion, then pass to assertion.
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewReader(byt))

		select {
		case test.requests <- *r:
			// do nothing, success
		case <-time.After(time.Second):
			require.Fail(t, "unexpected executor request")
		}

		// Forward this request on to the SDK.
		url := sdkURL
		url.RawQuery = r.URL.RawQuery
		req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(byt))
		require.NoError(t, err)
		req.ContentLength = r.ContentLength
		req.Header = r.Header
		sdkResponse, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		// Read the response.
		byt, err = io.ReadAll(sdkResponse.Body)
		require.NoError(t, err)
		sdkResponse.Body.Close()
		sdkResponse.Body = ioutil.NopCloser(bytes.NewReader(byt))

		fmt.Printf(" ==> Received SDK response:\n\t%s\n", string(byt))
		fmt.Println("")

		select {
		case test.responses <- *sdkResponse:
			// do nothing, success
		case <-time.After(time.Second):
			require.Fail(t, "unexpected sdk response")
		}

		// Forward the response from the SDK to the executor.
		w.WriteHeader(sdkResponse.StatusCode)
		_, err = w.Write(byt)
		require.NoError(t, err)
	}))
	srv := &http.Server{
		Addr:    ":42018",
		Handler: mux,
	}
	go srv.ListenAndServe()
	defer srv.Close()

	localURL, err := url.Parse(proxyURL)
	require.NoError(t, err)

	// Register all functions with the SDK.
	err = register(*localURL, *rr)
	require.NoError(t, err, "Function registration error")

	// Trigger the function by sending an event.
	trigger := test.Function.Triggers[0]
	if trigger.EventTrigger == nil {
		t.Fatalf("Unable to trigger scheduled functions")
		return
	}

	test.test = t
	for _, f := range test.chain {
		f()
	}

	fmt.Println(" ==> Waiting for extraneous requests")
	<-time.After(test.Timeout + buffer)
	fmt.Printf("\n\n")
}

// introspect asserts that the SDK is live and the expected function exists when calling
// the introspect handler.
func introspect(test *Test) (*sdk.RegisterRequest, error) {
	url := sdkURL
	url.Path = "/api/inngest"
	url.RawQuery = "introspect"

	resp, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("unable to call introspect on SDK: %w", err)
	}
	defer resp.Body.Close()

	data := &sdk.RegisterRequest{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid introspect response: unable to decode introspect response: %w", err)
	}

	expected, err := json.MarshalIndent(test.Function, "", "  ")
	if err != nil {
		return nil, err
	}

	found := false
	for _, f := range data.Functions {
		actual, _ := json.MarshalIndent(f, "", "  ")
		if bytes.Equal(expected, actual) {
			found = true
			break
		}
	}

	response, _ := json.MarshalIndent(data, "", "  ")
	if !found {
		return nil, fmt.Errorf("Expected function not found:\n%s\n\nIntrospection:\n%s", string(expected), string(response))
	}

	return data, nil
}

func register(serverURL url.URL, rr sdk.RegisterRequest) error {
	// Register functions using _this_ host and the introspection request
	for n, fn := range rr.Functions {
		for key, step := range fn.Steps {
			rt := step.Runtime.Runtime.(inngest.RuntimeHTTP)
			// Take the URL and replace the host with our server's URL.
			parsed, err := url.Parse(rt.URL)
			if err != nil {
				return err
			}
			serverURL.Path = "/"
			serverURL.RawQuery = parsed.RawQuery
			rt.URL = serverURL.String()
			step.Runtime.Runtime = rt
			fn.Steps[key] = step
		}
		rr.Functions[n] = fn
	}

	// Randomize the hash.
	hash := uuid.New().String()
	rr.Hash = &hash

	byt, err := json.Marshal(rr)
	if err != nil {
		return err
	}

	url := apiURL
	url.Path = "/fn/register"
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(byt))
	if err != nil {
		return err
	}

	key := regexp.MustCompile(`^signkey-[\w]+-`).ReplaceAllString(signingKey, "")
	byt, _ = hex.DecodeString(key)
	sum := sha256.Sum256(byt)
	keyHash := hex.EncodeToString(sum[:])

	req.Header.Add("Authorization", fmt.Sprintf("Bearer signkey-test-%s", keyHash))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error registering: %w", err)
	}
	if resp.StatusCode > 299 {
		byt, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("Error when registering functions: %s", string(byt))
	}

	return nil
}

func stepURL(fnID string, step string) string {
	url := sdkURL
	url.Path = "/api/inngest"
	q := url.Query()
	q.Add("fnId", fnID)
	q.Add("stepId", step)
	url.RawQuery = q.Encode()
	return url.String()
}
