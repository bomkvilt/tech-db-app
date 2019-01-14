package models

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// easyjson:json
type Post struct {
	ID       int64     `json:"id,omitempty"`
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	Forum    string    `json:"forum,omitempty"`
	IsEdited bool      `json:"isEdited,omitempty"`
	Message  string    `json:"message"`
	Parent   int64     `json:"parent,omitempty"`
	Thread   int       `json:"thread,omitempty"`
	Path     []int64   `json:"-"`
}

// easyjson:json
type Posts []*Post

// easyjson:json
type PostFull struct {
	Author *User   `json:"author,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
	Post   *Post   `json:"post,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
}

// easyjson:json
type PostQuery struct {
	Desc     *bool
	Limit    *int
	Since    int64
	SlugOrID string
	Sort     string
}

func (Post) FromRequest(r *http.Request) *Post {
	b, err := ioutil.ReadAll(r.Body)
	check(err)

	u := &Post{}
	check(u.UnmarshalJSON(b))
	return u
}

func (Posts) FromRequest(r *http.Request) Posts {
	b, err := ioutil.ReadAll(r.Body)
	check(err)

	u := Posts{}
	check(u.UnmarshalJSON(b))
	return u
}

func (PostQuery) FromRequest(r *http.Request) *PostQuery {
	var (
		limit = r.URL.Query().Get("limit")
		since = r.URL.Query().Get("since")
		desc  = r.URL.Query().Get("desc")
		sort  = r.URL.Query().Get("sort")
		q     = &PostQuery{}
	)
	q.SlugOrID = mux.Vars(r)["slug_or_id"]
	if limit != "" {
		q.Limit = new(int)
		*q.Limit, _ = strconv.Atoi(limit)
	}
	if since != "" {
		q.Since, _ = strconv.ParseInt(since, 10, 64)
	}
	if desc != "" {
		q.Desc = new(bool)
		if desc == "true" {
			*q.Desc = true
		}
	}
	if sort != "" {
		q.Sort = sort
	} else {
		q.Sort = "flat"
	}
	return q
}
