package sessions

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestSessions(T *testing.T) {

	rand.Seed(1337)

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
			MaxAge: rand.Intn(10),
		}
		s = New(mux, c, timeoutFunc, sessionskey{})
		http.ListenAndServe(":12345", s)
	}()

	f := func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

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
	}
	for i := 0; i < 100; i++ {
		go f()
	}

	<-time.After(10 * time.Second)
}
