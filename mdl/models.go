package mdl

type Community struct {
	Id          int    `db:"community_id, primarykey, autoincrement"`
	Key         string `db:"community_key, size:200, not null"` // unique with SetUnique
	Name        string `db:"community_name, size:200, not null"`
	PhoneNumber string `db:"meta_phonenumber, size:200, not null"`
}
