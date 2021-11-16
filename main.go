package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type PostID int

type Post struct {
	ID      PostID
	Content string
	Replies []*Post
}

var nextID PostID = 1

var root = Post{
	ID:      0,
	Content: "Welcome to the forum!",
	Replies: []*Post{},
}

var rootMutex sync.Mutex

var templates = template.Must(template.ParseFiles("templates/post.html", "templates/index.html"))

func main() {
	insertReply(&root, "first")

	http.HandleFunc("/reply/", replyHandler)
	http.HandleFunc("/", indexHandler)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	fmt.Println("listening on http://localhost:8000")
	http.ListenAndServe(":8000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", &root)
}

func replyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/reply/"), 10, 32)
	if err != nil {
		fmt.Fprintln(w, "invalid id", err)
		return
	}

	if post := findPost(PostID(id), &root); post != nil {
		insertReply(post, r.URL.Query().Get("content"))
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		fmt.Fprintln(w, "couldn't find post", id)
	}
}

func findPost(id PostID, root *Post) *Post {
	// can't lock mutex here

	if root.ID == id {
		return root
	}

	for _, child := range root.Replies {
		if p := findPost(id, child); p != nil {
			return p
		}
	}

	return nil
}

func insertReply(parent *Post, content string) {
	rootMutex.Lock()
	defer rootMutex.Unlock()

	parent.Replies = append(parent.Replies, &Post{
		ID:      nextID,
		Content: content,
		Replies: []*Post{},
	})

	nextID++
}
