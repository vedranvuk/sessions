sessions
========

Sessions implements http.Handler interface and acts as a filter between the actual handler and a HTTP server. It manages session cookies and provides a simple mechanism for extracting a session id from a http.Request.

	func Test() {
	
		go func() {
			type sessionskey struct{} // Unique context key for session id.
			var s *Sessions // The sessions object.
			
			// Function called when a session times out.
			timeoutFunc := func(sessionid int) {
				fmt.Printf("session %d timed out.\r\n", sessionid)
			}
	
			// An example handler.
			handlerFunc := func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "session id: %d\r\n", s.SID(r))
			}
	
			// Actual mux and handler registration.
			mux := http.NewServeMux()
			mux.HandleFunc("/", handlerFunc)
	
			// Template session coookie.
			c := &http.Cookie{
				Name:   "sessionstest", // Name of session cookies.
				MaxAge: 5, // Cookie age in seconds.
			}
			
			// Create a new Session passing all the stuff into constructor.
			s = New(mux, c, timeoutFunc, sessionskey{})
			
			// RUn the server.
			http.ListenAndServe(":12345", s)
		}()
	
		// Perform a get and display session id.
		r, _ := http.Get("http://localhost:12345/")
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(string(b))
	
		<-time.After(10 * time.Second)
	}