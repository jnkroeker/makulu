package schema

// UploadFeedRequest is the data required to make a feed/upload request
type UploadFeedRequest struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// UploadFeedResponse is the response from the feed/upload request
type UploadFeedResponse struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Message string  `json:"message"`
}
