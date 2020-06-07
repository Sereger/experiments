package nested

type Doc struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Items []*Doc `json:"items,omitempty"`
}
