package wirelesstag

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/oauth2"
)

func TestExchangeToken(t *testing.T) {
	Convey("Exchange", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"access_token": "foo", "token_type": "Bearer", "expiry": "1970-01-01T00:00:00Z"}`)
		}))
		defer ts.Close()

		config := &oauth2.Config{
			Endpoint: oauth2.Endpoint{
				TokenURL: ts.URL,
			},
		}
		token, err := exchangeToken(config, "foobar")
		So(err, ShouldBeNil)
		So(token.AccessToken, ShouldEqual, "foo")
		So(token.TokenType, ShouldEqual, "Bearer")
		So(token.Expiry, ShouldEqual, time.Unix(0, 0))
	})
}
