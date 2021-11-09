package offchain

import (
<<<<<<< HEAD
	"fmt"
=======
>>>>>>> development
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const defaultTestURI = "http://example.url"

func TestHTTPSetLimit(t *testing.T) {
	t.Parallel()

	set := NewHTTPSet()
	var err error
	for i := 0; i < maxConcurrentRequests+1; i++ {
		_, err = set.StartRequest(http.MethodGet, defaultTestURI)
	}

	require.ErrorIs(t, errIntBufferEmpty, err)
}

func TestHTTPSet_StartRequest_NotAvailableID(t *testing.T) {
	t.Parallel()

	set := NewHTTPSet()
	set.reqs[1] = &OffchainRequest{}

	_, err := set.StartRequest(http.MethodGet, defaultTestURI)
	require.ErrorIs(t, errRequestIDNotAvailable, err)
}

func TestHTTPSetGet(t *testing.T) {
	t.Parallel()

	set := NewHTTPSet()

	id, err := set.StartRequest(http.MethodGet, defaultTestURI)
	require.NoError(t, err)

	req := set.Get(id)
	require.NotNil(t, req)

	require.Equal(t, http.MethodGet, req.Request.Method)
	require.Equal(t, defaultTestURI, req.Request.URL.String())
}

func TestOffchainRequest_AddHeader(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		offReq           OffchainRequest
		err              error
		headerK, headerV string
	}{
		"should return invalid request": {
			offReq: OffchainRequest{invalid: true},
			err:    errInvalidRequest,
		},
		"should return request already started": {
			offReq: OffchainRequest{waiting: true},
			err:    errRequestAlreadyStarted,
		},
		"should add header": {
			offReq:  OffchainRequest{Request: &http.Request{Header: make(http.Header)}},
			headerK: "key",
			headerV: "value",
		},
		"should return invalid empty header": {
			offReq:  OffchainRequest{Request: &http.Request{Header: make(http.Header)}},
			headerK: "",
			headerV: "value",
			err:     fmt.Errorf("%w: %s", errInvalidHeaderKey, "empty header key"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.offReq.AddHeader(tc.headerK, tc.headerV)

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), err.Error())
				return
			}

			got := tc.offReq.Request.Header.Get(tc.headerK)
			require.Equal(t, tc.headerV, got)
		})
	}
}