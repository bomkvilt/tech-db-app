package database

import (
	"AForum/internal/models"
	"strconv"
)

// CreateNewThread -
func (m *DB) CreateNewThread(forum string, th *models.Thread) (err error) {
	var (
		slug    = &th.Slug
		o, err0 = m.GetThreadBySlug(th.Slug)
		u, err1 = m.GetUserByName(th.Author)
		f, err2 = m.GetForumBySlug(forum)
	)
	if err0 == nil {
		*th = *o
		return AlreadyExist(nil)
	}
	for _, err := range []error{err1, err2} {
		if err != nil {
			return NotFound(err)
		}
	}
	if *slug == "" {
		slug = nil
	}
	th.Author = u.Nickname
	th.Forum = f.Slug

	tx := m.db.MustBegin()
	defer tx.Rollback()
	if err := tx.QueryRow(`
		INSERT INTO threads(author, created, forum, message, slug, title, votes)
		VALUES ($1, $2, $3, $4, $5, $6, 0)
		RETURNING tid
	`, u.ID, th.Created, f.ID, th.Message, slug, th.Title).Scan(&th.ID); err != nil {
		return AlreadyExist(err)
	}
	if _, err := tx.Exec(`
		UPDATE forums
		SET threadCount=threadCount+1
		WHERE fid=$1
	`, f.ID); err != nil {
		return NotFound(err)
	}
	return tx.Commit()
}

func (m *DB) UpdateThread(slugOrID string, th *models.Thread) (err error) {
	if _, err = m.GetThreadBySlugOrID(slugOrID); err != nil {
		return NotFound(err)
	}

	tx := m.db.MustBegin()
	defer tx.Rollback()
	if _, err := tx.Exec(`
		UPDATE threads
		SET title  =COALESCE(NULLIF($1,''), title  ),
			message=COALESCE(NULLIF($2,''), message)
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

	tx := m.db.MustBegin()
	defer tx.Rollback()
	r, _ := tx.Exec(`
		UPDATE votes
		SET voice=$3
		WHERE thread=$1 AND author=$2;
	`, th.ID, u.ID, vt.Voice)
	if num, _ := r.RowsAffected(); num != 1 {
		if _, err := tx.Exec(`
			INSERT INTO votes(thread, author, voice)
			VALUES ($1, $2, $3);
		`, th.ID, u.ID, vt.Voice); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(`
		UPDATE threads
		SET votes=(SELECT SUM(voice) FROM votes WHERE thread=$1)
		WHERE tid=$1
	`, th.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return m.GetThreadBySlugOrID(slugOrID)
}

func (m *DB) getThread(key string, value interface{}) (thr *models.Thread, err error) {
	thr = &models.Thread{}
	if err := m.db.QueryRow(`
		SELECT u.nickname, t.created, f.slug, t.tid, t.message, t.slug, t.title, t.votes
		FROM threads t
		JOIN users  u ON t.author=u.uid
		JOIN forums f ON t.forum=f.fid
		WHERE t.`+key+`=$1
	`, value).
		Scan(&thr.Author, &thr.Created, &thr.Forum, &thr.ID, &thr.Message, &thr.Slug, &thr.Title, &thr.Votes); err != nil {
		return nil, err
	}
	return thr, nil
}

func (m *DB) GetThreadByID(tid int) (thr *models.Thread, err error) { return m.getThread("tid", tid) }

func (m *DB) GetThreadBySlug(slug string) (thr *models.Thread, err error) {
	if slug == "" {
		return nil, NotFound(nil)
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
	if q.Since != nil {
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

	for rows.Next() {
		tmp := &models.Thread{Forum: f.Slug}
		if err := rows.Scan(&tmp.Author, &tmp.Created, &tmp.ID, &tmp.Message, &tmp.Slug, &tmp.Title, &tmp.Votes); err != nil {
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}