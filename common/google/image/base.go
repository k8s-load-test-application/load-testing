package image

type Image struct {
	// These are image metadata which are fetched from google search
	Url    string `json:"url"`
	Title  string `json:"title"`
	Base   string `json:"base"`
	Source string `json:"source"`
}
