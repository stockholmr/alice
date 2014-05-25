// Package alice implements a middleware chaining solution.
package alice

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A constructor for middleware
// that writes its own "tag" into the RW and does nothing else.
// Useful in checking if a chain is behaving in the right order.
func tagMiddleware(tag string) Constructor {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(tag))
			h.ServeHTTP(w, r)
		})
	}
}

// Tests creating a new chain
func TestNew(t *testing.T) {
	c1 := func(h http.Handler) http.Handler {
		return nil
	}
	c2 := func(h http.Handler) http.Handler {
		return http.StripPrefix("potato", nil)
	}

	slice := []Constructor{c1, c2}

	chain := New(slice...)
	assert.Equal(t, chain.constructors[0], slice[0])
	assert.Equal(t, chain.constructors[1], slice[1])
}

func TestThenWorksWithNoMiddleware(t *testing.T) {
	assert.NotPanics(t, func() {
		New()
	})
}

func TestThenTreatsNilAsDefaultServeMux(t *testing.T) {
	chained := New().Then(nil)
	assert.Equal(t, chained, http.DefaultServeMux)
}

func TestThenOrdersHandlersRight(t *testing.T) {
	t1 := tagMiddleware("t1\n")
	t2 := tagMiddleware("t2\n")
	t3 := tagMiddleware("t3\n")
	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("app\n"))
	})

	chained := New(t1, t2, t3).Then(app)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	chained.ServeHTTP(w, r)

	assert.Equal(t, w.Body.String(), "t1\nt2\nt3\napp\n")
}