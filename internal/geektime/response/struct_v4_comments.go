package response

// V4CommentListResponse ...
type V4CommentListResponse struct {
	Data struct {
		Page struct {
			HasMore bool `json:"more"`
			Count int `json:"count"`
		} `json:"page"`

		CommentList []struct {
			UserName string `json:"user_name"`
			CommentCtime int `json:"comment_ctime"`
			Replies []struct {
				UserName string `json:"user_name"`
				Content string `josn:"content"`
				CTime int `json:"ctime"`
			} `json:"replies"`
			CommentContent string `json:"comment_content"`
		} `json:"list"`

	} `json:"data"`
	Code int `json:"code"`
}
