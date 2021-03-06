package database

import (
	"AForum/internal/models"
	"errors"
	"strconv"
)

// CreateNewThread -
func (m *DB) CreateNewThread(forum string, th *models.Thread) (err error) {
	var (
		u, err1 = m.GetUserByName(th.Author)
		f, err2 = m.GetForumBySlug(forum)
	)
	if o, err := m.GetThreadBySlug(th.Slug); err == nil {
		*th = *o
		return AlreadyExist(errors.New("Thread already exists"))
	}
	for _, err := range []error{err1, err2} {
		if err != nil {
			return NotFound(err)
		}
	}
	th.Author = u.Nickname
	th.Forum = f.Slug

	tx, _ := m.db.Begin()
	defer tx.Rollback()
	if err := tx.QueryRow(`
		INSERT INTO threads(author, created, forum, message, slug          , title , votes)
		VALUES             ($1    , $2     , $3   , $4     , NULLIF($5, ''), $6    , 0    )
		RETURNING tid
	`, u.ID, th.Created, f.ID, th.Message, th.Slug, th.Title).Scan(&th.ID); err != nil {
		return AlreadyExist(err)
	}
	return tx.Commit()
}

func (m *DB) UpdateThread(slugOrID string, th *models.Thread) (err error) {
	if o, err := m.GetThreadBySlugOrID(slugOrID); err != nil {
		return NotFound(err)
	} else {
		th.ID = o.ID
	}

	tx, _ := m.db.Begin()
	defer tx.Rollback()
	if _, err := tx.Exec(`
		UPDATE threads
		SET title   =COALESCE(NULLIF($1,''), title  ),
			message =COALESCE(NULLIF($2,''), message)
		WHERE tid=$3
	`, th.Title, th.Message, th.ID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	o, err := m.GetThreadBySlugOrID(slugOrID)
	*th = *o
	return err
}

// VoteThread -
func (m *DB) VoteThread(slugOrID string, vt *models.Vote) (*models.Thread, error) {
	var (
		th, err1 = m.GetThreadBySlugOrID(slugOrID)
		u, err2  = m.GetUserByName(vt.Nickname)
	)
	for _, err := range []error{err1, err2} {
		if err != nil {
			return nil, NotFound(err)
		}
	}

	tx, _ := m.db.Begin()
	defer tx.Rollback()
	r, err := tx.Exec(`
		UPDATE votes
		SET voice=$3
		WHERE thread=$1 AND author=$2;
	`, th.ID, u.ID, vt.Voice)
	check(err)

	// new vote
	if r.RowsAffected() == 0 {
		if _, err := tx.Exec(`
			INSERT INTO votes(thread, author, voice)
			VALUES ($1, $2, $3);
		`, th.ID, u.ID, vt.Voice); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return m.GetThreadBySlugOrID(slugOrID)
}

func (m *DB) getThread(key string, value interface{}) (t *models.Thread, err error) {
	t = &models.Thread{}
	s := new(string)
	if err := m.db.QueryRow(`
		SELECT u.nickname, t.created, f.slug, t.tid, t.message, t.slug, t.title, t.votes
		FROM       threads t
		INNER JOIN users   u ON t.author=u.uid
		INNER JOIN forums  f ON t.forum =f.fid
		WHERE t.`+key+`=$1
	`, value).
		Scan(&t.Author, &t.Created, &t.Forum, &t.ID, &t.Message, &s, &t.Title, &t.Votes); err != nil {
		return nil, err
	}
	if s != nil {
		t.Slug = *s
	}
	return t, nil
}

func (m *DB) GetThreadByID(tid int) (thr *models.Thread, err error) { return m.getThread("tid", tid) }

func (m *DB) GetThreadBySlug(slug string) (thr *models.Thread, err error) {
	if slug == "" {
		return nil, NotFound(errors.New("Empty thread slug"))
	}
	return m.getThread("slug", slug)
}

func (m *DB) GetThreadBySlugOrID(slugOrID string) (thr *models.Thread, err error) {
	if tid, err := strconv.Atoi(slugOrID); err == nil {
		return m.GetThreadByID(tid)
	}
	return m.GetThreadBySlug(slugOrID)
}

// GetThreads -
func (m *DB) GetThreads(q *models.ForumQuery) (res models.Threads, err error) {
	f, err := m.GetForumBySlug(q.Slug)
	if err != nil {
		return nil, NotFound(err)
	}

	var (
		order = "ASC"
		vars  = make([]interface{}, 1, 3)
		parts = make(map[string]string)
	)
	vars[0] = &f.ID
	if q.Desc != nil && *q.Desc {
		order = "DESC"
	}
	if q.Since != "" {
		sign := ">="
		if order == "DESC" {
			sign = "<="
		}
		parts["since"] = "AND t.created" + sign + "$" + strconv.Itoa(len(vars)+1)
		vars = append(vars, q.Since)
	}
	if q.Limit != nil {
		parts["limit"] = "LIMIT $" + strconv.Itoa(len(vars)+1)
		vars = append(vars, q.Limit)
	}

	rows, err := m.db.Query(`
		SELECT u.nickname, t.created, t.tid, t.message, t.slug, t.title, t.votes
		FROM threads t
		JOIN users   u ON t.author=u.uid
		WHERE t.forum = $1 `+parts["since"]+`
		ORDER BY t.created `+order+`
		`+parts["limit"]+`
	`, vars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res = models.Threads{}
	for rows.Next() {
		s := new(string)
		t := &models.Thread{Forum: f.Slug}
		if err := rows.Scan(&t.Author, &t.Created, &t.ID, &t.Message, &s, &t.Title, &t.Votes); err != nil {
			return nil, err
		}
		if s != nil {
			t.Slug = *s
		}
		res = append(res, t)
	}
	return res, nil
}
