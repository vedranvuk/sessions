// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sessions implements a http.Handler filter that acts as a session
// manager and transparently layers session cookies over requests and responses.
package sessions

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/vedranvuk/timedlist"
)

// Session timeout event handler function prototype.
// Carries a single parameter, integer id of a session that timed out.
type TimeoutFunc func(int)

// Sessions implements the http.Handler interface and acts as a filter between a
// "real" handler and a HTTP server. It should prefferably be the first in the
// otherwise not recommended chain of handlers.
type Sessions struct {
	tmpl *http.Cookie
	tf   TimeoutFunc
	list *timedlist.TimedList
	next http.Handler
	key  interface{}
}

// New creates a new instance of Sessions.
// next is the http.Handler that ServeHTTP gets forwarded to.
// tmpl is a cookie that is used as a template for session cookies. The
// following values are interpreted, all others are copied directly to new
// session cookies:
//	Name: Session cookie name.
// 	MaxAge: Sets session length in seconds.
//  Value: Ignored. Value is set to session ID and is managed internally.
//	Expires: Ignored. Calculated from MaxAge.
// tf is a function that gets called when a session expires.
// key is any type supported as key by Context package.
func New(next http.Handler, tmpl *http.Cookie, tf TimeoutFunc, key interface{}) *Sessions {
	if tmpl == nil {
		tmpl = &http.Cookie{
			Name:   "sessionid",
			MaxAge: 180,
			Path:   "/",
		}
	}
	p := &Sessions{
		tmpl: tmpl,
		tf:   tf,
		key:  key,
		next: next,
	}
	p.list = timedlist.New(p.timeout)
	return p
}

// Timedlist timeout event handler.
func (s *Sessions) timeout(p *timedlist.Item) {
	if s.tf == nil {
		return
	}
	s.tf(0)
}

// Makes a new session cookie from the template cookie.
func (s *Sessions) newCookie() *http.Cookie {
	c := &http.Cookie{}
	*c = *s.tmpl
	c.Value = "-1"
	return c
}

// SID returns a session id from a request.
func (s *Sessions) SID(r *http.Request) int {
	return r.Context().Value(s.key).(int)
}

// Process reads the r for session cookies, updates session or creates a new one
// and finally makes a single modification of w by adding a new Set-Cookie
// header with the value of a session cookie.
func (s *Sessions) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get or make a new session cookie.
	sid := -1
	c, e := r.Cookie(s.tmpl.Name)
	if e != nil {
		c = s.newCookie()
	}
	if n, e := strconv.Atoi(c.Value); e == nil {
		sid = n
	}
	c.Expires = time.Now().Add(time.Duration(s.tmpl.MaxAge) * time.Second)

	// Create a new or renew an existing session.
	d := time.Duration(s.tmpl.MaxAge) * time.Second
	s.list.MuLock()
	if s.list.MuExists(sid) {
		s.list.MuRenew(sid, d)
	} else {
		p := s.list.MuAdd(nil, d)
		sid = p.Id()
		c.Value = strconv.Itoa(sid)
	}
	s.list.MuUnlock()

	http.SetCookie(w, c)
	ctx := context.WithValue(r.Context(), s.key, sid)
	s.next.ServeHTTP(w, r.WithContext(ctx))
}
