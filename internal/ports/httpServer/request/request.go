package request

type UrlRequest struct{
	URL       string   `json:"url" binding:"required"`
	//Expiry string   `json:"expiry"`
}