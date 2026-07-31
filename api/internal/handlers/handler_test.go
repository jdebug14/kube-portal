package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jdebug14/kube-portal/internal/k8s"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func setupHandler(t *testing.T, objects ...runtime.Object) (*Handler, *fake.Clientset) {
	t.Helper()
	fakeCS := fake.NewSimpleClientset(objects...)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	k8sClient := k8s.NewClientFromInterface(fakeCS, logger)
	h := NewHandler(k8sClient, slog.Default())
	return h, fakeCS
}

func newRequestWithParams(t *testing.T, method, url string, rParams map[string]string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, url, nil)
	rctx := chi.NewRouteContext()
	for k, v := range rParams {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestValidateNamespaceName(t *testing.T) {
	tests := []struct {
		caseName  string
		name      string
		expectErr bool
	}{
		{caseName: "only alphanumeric", name: "mynamespace"},
		{caseName: "with numbers", name: "mynam35pace"},
		{caseName: "with hypens", name: "my-namespace"},
		{caseName: "with numbers and hyphens", name: "my-nam35pace"},
		{caseName: "only numbers", name: "123"},
		{caseName: "start with number", name: "123mynamespace"},
		{caseName: "end with number", name: "mynamepace123"},
		{caseName: "one", name: "a"},
		{caseName: "two", name: "ab"},
		{caseName: "exactly at cap", name: strings.Repeat("a", 63)},
		{caseName: "over cap", name: strings.Repeat("a", 64), expectErr: true},
		{caseName: "uppercase", name: "myNamespace", expectErr: true},
		{caseName: "whitespace", name: "my namespace", expectErr: true},
		{caseName: "empty string", name: "", expectErr: true},
		{caseName: "start with hyphen", name: "-mynamespace", expectErr: true},
		{caseName: "end with hyphen", name: "mynamespace-", expectErr: true},
		{caseName: "underscore", name: "my_namespace", expectErr: true},
		{caseName: "dot", name: "my.namespace", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.caseName, func(t *testing.T) {
			err := validateNamespaceName(tc.name)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateResourceName(t *testing.T) {
	tests := []struct {
		caseName  string
		name      string
		expectErr bool
	}{
		{caseName: "only alphanumeric", name: "myresource"},
		{caseName: "with numbers", name: "myr35ource"},
		{caseName: "with hypens", name: "my-resource"},
		{caseName: "with numbers and hyphens", name: "my-r35ource"},
		{caseName: "only numbers", name: "123"},
		{caseName: "start with number", name: "123myresource"},
		{caseName: "end with number", name: "myresource123"},
		{caseName: "one", name: "a"},
		{caseName: "two", name: "ab"},
		{caseName: "exactly at cap", name: strings.Repeat("a", 253)},
		{caseName: "over cap", name: strings.Repeat("a", 254), expectErr: true},
		{caseName: "uppercase", name: "myResource", expectErr: true},
		{caseName: "whitespace", name: "my resource", expectErr: true},
		{caseName: "empty string", name: "", expectErr: true},
		{caseName: "start with hyphen", name: "-myresource", expectErr: true},
		{caseName: "end with hyphen", name: "myresource-", expectErr: true},
		{caseName: "underscore", name: "my_resource", expectErr: true},
		{caseName: "dot", name: "my.resource"},
	}

	for _, tc := range tests {
		t.Run(tc.caseName, func(t *testing.T) {
			err := validateResourceName(tc.name)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
