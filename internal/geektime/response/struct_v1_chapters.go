package response

type V1ChaptersResponse struct {
	Data  []struct {
		SourceID         int       `json:"source_id"`
		Title            string       `json:"title"`
		ArticleCount     int          `json:"article_count"`
		ID               string       `json:"id"`
	} `json:"data"`
	Code int `json:"code"`
}