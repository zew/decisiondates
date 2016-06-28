package mdl

type Community struct {
	Id   int    `db:"community_id, primarykey, autoincrement"`
	Key  string `db:"community_key, size:200, not null"` // unique with SetUnique
	Name string `db:"community_name, size:200, not null"`
	// PhoneNumber string `db:"meta_phonenumber, size:200, not null"`
}

type Pdf struct {
	Id       int    `db:"pdf_id, primarykey, autoincrement"`
	Key      string `db:"community_key, size:200, not null"`
	Name     string `db:"community_name, size:200, not null"`
	Url      string `db:"pdf_url, size:200, not null"` // SetUniqueTogether(community_key, pdf_url )
	Content  string `db:"pdf_test, size:4200, not null"`
	Snippet1 string `db:"pdf_snippet1, size:400, not null"`
	Snippet2 string `db:"pdf_snippet2, size:400, not null"`
	Snippet3 string `db:"pdf_snippet3, size:400, not null"`
}
