package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type FeedItem struct {
	ID                 int
	Header, Body, Tags string
}

type Server interface {
	Feed(http.ResponseWriter, *http.Request)
	Add(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
	ClientSideFeed(http.ResponseWriter, *http.Request)
	Shutdown()
}

func NewServer(ctx context.Context, filename string) (Server, error) {
	log.Printf("Opening database in `%s`", filename)
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	err = initDatabase(ctx, database)
	if err != nil {
		return nil, err
	}

	return &server{
		db: database,
	}, nil
}

type server struct {
	db *sql.DB
}

func initDatabase(ctx context.Context, db *sql.DB) error {
	log.Print("Initializing database")
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, header TEXT, body TEXT, tags TEXT)
`)
	return err
}

func randomString() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *server) ClientSideFeed(w http.ResponseWriter, r *http.Request) {
	err := writeHead(w)
	if err != nil {
		return
	}
	w.Write([]byte(fmt.Sprintf(`
<script>
const data = %s;
data.forEach(item => {
  const container = document.createElement('p');
  const header = document.createElement('h1');
  const body = document.createElement('p');
  header.innerText = item.Header;
  body.innerText = item.Body;
  container.appendChild(header);
  container.appendChild(body);
  document.body.appendChild(container);
});
</script>
`, s.renderFeedAsData(r))))

	_ = writeTrailer(w)
}

func (s *server) renderFeedAsData(r *http.Request) string {
	feed, err := s.getFeed(r.Context(), r.URL.Query().Get("tag"), r.URL.Query().Get("header"))
	if err != nil {
		return "[]"
	}

	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(feed)
	if err != nil {
		return "[]"
	}
	return string(buffer.Bytes())
}

func (s *server) Feed(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "extremely_secret_cookie_never_show",
		Value: randomString(),
	})
	log.Print("Rendering the feed...")
	err := s.writeFeed(r.Context(), w, r.URL.Query().Get("tag"), r.URL.Query().Get("header"))
	if err != nil {
		log.Print(err)
	}
}

func (s *server) Add(w http.ResponseWriter, r *http.Request) {
	log.Print("Posting the feed...")
	ctx := r.Context()
	r.ParseForm()
	query := "INSERT INTO posts (header, body, tags) VALUES (?, ?, ?)"
	_, err := s.db.ExecContext(ctx, query, r.Form.Get("header"), r.Form.Get("body"), r.Form.Get("tags"))
	if err != nil {
		http.Error(w, "Failed to add the post", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/feed", http.StatusFound)
}

func (s *server) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.URL.Query().Get("id")
	log.Printf("Removing post with ID=%s", id)
	query := fmt.Sprintf("DELETE FROM posts WHERE id=%s", id)
	log.Printf("Executing query:\n%s\n", query)
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		http.Error(w, "Failed to remove the post", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/feed", http.StatusFound)
}

func (s *server) Shutdown() {
	s.db.Close()
}

func (s *server) getFeed(ctx context.Context, tag, header string) (result []FeedItem, err error) {
	query := "SELECT id, header, body, tags FROM posts"
	where := []string{}
	if tag != "" {
		where = append(where, "tags LIKE '%"+tag+"%'")
	}
	if header != "" {
		where = append(where, "header='"+header+"'")
	}
	if len(where) != 0 {
		query = query + " WHERE " + strings.Join(where, " AND ")
	}
	log.Printf("Executing query:\n%s\n", query)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	result = make([]FeedItem, 0, 100)
	for rows.Next() {
		item := FeedItem{}
		err = rows.Scan(&item.ID, &item.Header, &item.Body, &item.Tags)
		if err != nil {
			break
		}
		result = append(result, item)
	}
	return result, err
}

func (s *server) writeFeed(ctx context.Context, w http.ResponseWriter, tag, header string) error {
	log.Print("Write feed")
	err := writeHead(w)
	if err != nil {
		return err
	}
	w.Write([]byte(`
<form>
  <label for="search">Search</label>
  <input id="search" type="text" name="header"/>
</form>
`))

	if tag != "" {
		w.Write([]byte(fmt.Sprintf("<h1>Display posts for tag `%s`</h1>", tag)))
	}

	if header != "" {
		w.Write([]byte(fmt.Sprintf("<h1>Posts with header `%s`</h1>", header)))
	}

	feed, err := s.getFeed(ctx, tag, header)
	if err != nil {
		return err
	}

	for _, item := range feed {
		w.Write([]byte(fmt.Sprintf(`
<h3 id="%d">%s</h3>
<p>%s</p>
<p>Tags:</p>
<p>%s</p>
<form action="/feed/delete">
  <input type="hidden" value="%d" name="id" />
  <button type="submit">Delete</button>
</form>
`, item.ID, item.Header, item.Body, prepareTags(item.Tags), item.ID)))
	}
	if len(feed) == 0 {
		w.Write([]byte("<p>No posts yet</p>"))
	}
	_ = writeForm(w)
	_ = writeTrailer(w)
	return err
}

func prepareTags(tags string) string {
	strs := strings.Split(tags, ",")
	result := ""
	for _, str := range strs {
		tag := strings.TrimSpace(str)
		result = result + fmt.Sprintf("\n<a href=\"/feed?tag=%s\">%s</a>", tag, tag)
	}
	return result
}

func writeHead(w http.ResponseWriter) error {
	log.Print("Write head")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
	<head>
	  <title>The most terrible app ever created</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
  </head>
  <body>
`))
	return err
}

func writeForm(w http.ResponseWriter) error {
	log.Print("Write form")
	_, err := w.Write([]byte(`
<form method="post">
  <p>
    <label for="header">Header</label>
    <input id="header" name="header" type="text"/>
  </p>
  <p>
    <label for="body">Body</label>
    <textarea id="body" name="body" rows="10" cols="80"></textarea>
  </p>
  <p>
    <label for="tags">Tags</label>
    <input id="tags" name="tags" type="text"/>
  </p>
  <button type="submit">Post!</button>
</form>
`))
	return err
}

func writeTrailer(w http.ResponseWriter) error {
	log.Print("Write trailer")
	_, err := w.Write([]byte(`
  </body>
</html>
`))
	return err
}
