package database

import (
	"AForum/internal/models"
	"strconv"
	"strings"
)

// InsertNewUser -
func (db *DB) InsertNewUser(u *models.User) error {
	tx := db.db.MustBegin()
	defer tx.Rollback()

	if _, err := db.db.Exec(`
		INSERT INTO users(nickname, fullname, about, email)
		VALUES ( $1, $2, $3, $4 );
	`, u.Nickname, u.Fullname, u.About, u.Email); err != nil {
		return err
	}
	return tx.Commit()
}

// GetAllCollisions -
func (db *DB) GetAllCollisions(u *models.User) (usrs models.Users, err error) {
	rows, err := db.db.Query(`
		SELECT nickname, fullname, about, email 
		FROM users u
		WHERE u.nickname=$1 OR u.email=$2;
	`, u.Nickname, u.Email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usrs = models.Users{}
	for rows.Next() {
		t := &models.User{}
		if err := rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email); err != nil {
			return nil, err
		}
		usrs = append(usrs, t)
	}
	return usrs, nil
}

func (db *DB) GetForumUsers(params *models.ForumQuery) (usrs models.Users, err error) {
	f, err := db.GetForumBySlug(params.Slug)
	if err != nil {
		return nil, NotFound(err)
	}

	var (
		order = "ASC"
		vars  = make([]interface{}, 1, 3)
		parts = make(map[string]string)
	)
	{ // set flags
		vars[0] = &f.ID
		if params.Desc != nil && *params.Desc {
			order = "DESC"
		}
		if params.Since != nil {
			sign := ">"
			if order == "DESC" {
				sign = "<"
			}

			u, err := db.GetUserByName(*params.Since)
			if err != nil {
				return usrs, nil
			}
			parts["since"] = "AND u.nickname" + sign + "$" + strconv.Itoa(len(vars)+1)
			vars = append(vars, u.Nickname)
		}
		if params.Limit != nil {
			parts["limit"] = "LIMIT $" + strconv.Itoa(len(vars)+1)
			vars = append(vars, params.Limit)
		}
	}

	rows, err := db.db.Query(`
		SELECT u.nickname, u.fullname, u.about, u.email
		FROM       forum_users x
		INNER JOIN users       u ON(u.uid=x.username)
		WHERE x.forum=$1    `+parts["since"]+`
		ORDER BY u.nickname `+order+`
		`+parts["limit"]+`
	`, vars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usrs = models.Users{}
	for rows.Next() {
		t := &models.User{}
		if err := rows.Scan(&t.Nickname, &t.Fullname, &t.About, &t.Email); err != nil {
			return nil, err
		}
		usrs = append(usrs, t)
	}
	return usrs, nil
}

// ----------------| Get

func (db *DB) getUser(field string, value interface{}) (u *models.User, err error) {
	u = &models.User{}
	if err := db.db.QueryRow(`
		SELECT nickname, fullname, about, email, uid 
		FROM users 
		WHERE `+field+`=$1;
	`, value).Scan(&u.Nickname, &u.Fullname, &u.About, &u.Email, &u.ID); err != nil {
		return nil, err
	}
	return u, nil
}
func (db *DB) GetUserByName(nick string) (*models.User, error)   { return db.getUser("nickname", nick) }
func (db *DB) GetUserByEmail(email string) (*models.User, error) { return db.getUser("email", email) }
func (db *DB) GetUserByID(uid int64) (*models.User, error)       { return db.getUser("uid", uid) }

func (db *DB) CheckUserByName(nick string) bool {
	dm := 0
	if err := db.db.QueryRow(`
		SELECT uid 
		FROM users 
		WHERE nickname=$1;`, nick).
		Scan(&dm); err != nil {
		return false
	}
	return true
}

func (db *DB) CheckUsersByName(nicks map[string]bool) bool {
	if len(nicks) == 0 {
		return true
	}

	arr := make([]string, 0, len(nicks))
	for n := range nicks {
		arr = append(arr, `'`+n+`'`)
	}
	res, err := db.db.Exec(`
		SELECT uid
		FROM users 
		WHERE nickname = ANY (ARRAY[` + strings.Join(arr, ", ") + `])
	`)
	check(err)
	rws, _ := res.RowsAffected()
	return rws == int64(len(arr))
}

// ----------------|

// UpdateUser -
func (db *DB) UpdateUser(u *models.User) error {
	tx := db.db.MustBegin()
	defer tx.Rollback()

	_, err := tx.Exec(`
		UPDATE users
		SET
			fullname = COALESCE(NULLIF($1, ''), fullname),
			email    = COALESCE(NULLIF($2, ''), email),
			about    = COALESCE(NULLIF($3, ''), about)
		WHERE nickname=$4
		RETURNING fullname, email, about
		`,
		u.Fullname, u.Email, u.About, u.Nickname)
	if err != nil {
		return err
	}
	return tx.Commit()
}
