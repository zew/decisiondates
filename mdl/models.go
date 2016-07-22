package mdl

type Community struct {
	Id       int    `db:"community_id, primarykey, autoincrement"`
	Key      string `db:"community_key, size:200, not null"` // unique with SetUnique
	Name     string `db:"community_name, size:200, not null"`
	NameOrig string `db:"community_orig_name, size:200, not null"`
	// PhoneNumber string `db:"meta_phonenumber, size:200, not null"`
}

type Pdf struct {
	Id            int    `db:"pdf_id, primarykey, autoincrement"`
	CommunityKey  string `db:"community_key, size:40, not null"`
	CommunityName string `db:"community_name, size:200, not null"`
	Url           string `db:"pdf_url, size:600, not null"` // SetUniqueTogether(community_key, pdf_url )
	Frequency     int    `db:"pdf_frequency, not null"`
	Title         string `db:"pdf_title, size:200, not null"` // SetUniqueTogether(community_key, pdf_url )
	ResultRank    int    `db:"pdf_resultrank, not null"`
	SnippetGoogle string `db:"pdf_snippet_google, size:400, not null"`
	// Content       string `db:"pdf_text, size:100200, not null"`
	Snippet1 string `db:"pdf_snippet1, size:4000, not null"`
	Snippet2 string `db:"pdf_snippet2, size:4000, not null"`
	Snippet3 string `db:"pdf_snippet3, size:4000, not null"`
}

type Page struct {
	Id      int    `db:"page_id, primarykey, autoincrement"`
	Url     string `db:"pdf_url, size:600, not null"`      // SetUniqueTogether(page_url, page_number)
	Number  int    `db:"page_number, not null"`            // SetUniqueTogether(page_url, page_number)
	Content string `db:"page_text, size:131072, not null"` // text (64k) does not suffice; make it mediumtext.
}

type PdfJoinPage struct {
	Pdf
	Page
}

type Decision struct {
	Id            int    `db:"decision_id, primarykey, autoincrement"`
	CommunityKey  string `db:"community_key, size:40, not null"`
	CommunityName string `db:"community_name, size:200, not null"`
	ForYear       int    `db:"decision_for_year, not null"`
	PageId        int    `db:"page_id, not null"`
}
