package mdl

type Community struct {
	Id   int    `db:"community_id, primarykey, autoincrement"`
	Key  string `db:"community_key, size:200, not null"` // unique with SetUnique
	Name string `db:"community_name, size:200, not null"`
	// PhoneNumber string `db:"meta_phonenumber, size:200, not null"`
}

type Pdf struct {
	Id            int    `db:"pdf_id, primarykey, autoincrement"`
	CommunityKey  string `db:"community_key, size:40, not null"`
	CommunityName string `db:"community_name, size:200, not null"`
	Url           string `db:"pdf_url, size:600, not null"` // SetUniqueTogether(community_key, pdf_url )
	Frequency     int    `db:"pdf_frequency, not null"`
	Title         string `db:"pdf_title, size:200, not null"` // SetUniqueTogether(community_key, pdf_url )
	SnippetGoogle string `db:"pdf_snippet_google, size:400, not null"`
	Content       string `db:"pdf_text, size:100200, not null"`
	Snippet1      string `db:"pdf_snippet1, size:400, not null"`
	Snippet2      string `db:"pdf_snippet2, size:400, not null"`
	Snippet3      string `db:"pdf_snippet3, size:400, not null"`
}
