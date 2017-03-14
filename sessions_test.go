package sessions

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestSessions(T *testing.T) {

	go func() {
		type sessionskey struct{}
		var s *Sessions

		timeoutFunc := func(sessionid int) {
			fmt.Printf("session %d timed out.\r\n", sessionid)
		}

		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "session id: %d\r\n", s.SID(r))
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", handlerFunc)

		c := &http.Cookie{
			Name:   "sessionstest",
			MaxAge: 5,
		}
		s = New(mux, c, timeoutFunc, sessionskey{})
		http.ListenAndServe(":12345", s)
	}()

	r, e := http.Get("http://localhost:12345/")
	if e != nil {
		T.Fatal("get failed", e)
	}
	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		T.Fatal("error reading get response", e)
	}
	r.Body.Close()
	fmt.Println(string(b))

	<-time.After(10 * time.Second)
}
