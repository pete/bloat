package mastodon

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
	"encoding/json"
)

type StatusPleroma struct {
	InReplyToAccountAcct string `json:"in_reply_to_account_acct"`
}

type ReplyInfo struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

type UTime time.Time // "Unreliable Time"
func (t *UTime) UnmarshalJSON(b []byte) error {
	var tm time.Time
	err := json.Unmarshal(b, &tm)
	if err != nil {
		return nil
	}
	*t = UTime(tm)
	return nil
}
func (t UTime) MarshalJSON() ([]byte, error) {
	tm := time.Time(t)
	return json.Marshal(tm)
}

// Status is struct to hold status.
type Status struct {
	ID                 string       `json:"id"`
	URI                string       `json:"uri"`
	URL                string       `json:"url"`
	Account            Account      `json:"account"`
	InReplyToID        interface{}  `json:"in_reply_to_id"`
	InReplyToAccountID interface{}  `json:"in_reply_to_account_id"`
	Reblog             *Status      `json:"reblog"`
	Content            string       `json:"content"`
	UCreatedAt          UTime    `json:"created_at,omitempty"`
	Emojis             []Emoji      `json:"emojis"`
	RepliesCount       int64        `json:"replies_count"`
	ReblogsCount       int64        `json:"reblogs_count"`
	FavouritesCount    int64        `json:"favourites_count"`
	Reblogged          interface{}  `json:"reblogged"`
	Favourited         interface{}  `json:"favourited"`
	Muted              interface{}  `json:"muted"`
	Sensitive          bool         `json:"sensitive"`
	SpoilerText        string       `json:"spoiler_text"`
	Visibility         string       `json:"visibility"`
	MediaAttachments   []Attachment `json:"media_attachments"`
	Mentions           []Mention    `json:"mentions"`
	Tags               []Tag        `json:"tags"`
	Card               *Card        `json:"card"`
	Application        Application  `json:"application"`
	Language           string       `json:"language"`
	Pinned             interface{}  `json:"pinned"`
	Bookmarked         bool         `json:"bookmarked"`
	Poll               *Poll        `json:"poll"`

	// Custom fields
	Pleroma       StatusPleroma          `json:"pleroma"`
	ShowReplies   bool                   `json:"show_replies"`
	IDReplies     map[string][]ReplyInfo `json:"id_replies"`
	IDNumbers     map[string]int         `json:"id_numbers"`
	RetweetedByID string                 `json:"retweeted_by_id"`
}

func (s *Status) CreatedAt() time.Time {
	return time.Time(s.UCreatedAt)
}

// Context hold information for mastodon context.
type Context struct {
	Ancestors   []*Status `json:"ancestors"`
	Descendants []*Status `json:"descendants"`
}

// Card hold information for mastodon card.
type Card struct {
	URL          string `json:"url"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Image        string `json:"image"`
	Type         string `json:"type"`
	AuthorName   string `json:"author_name"`
	AuthorURL    string `json:"author_url"`
	ProviderName string `json:"provider_name"`
	ProviderURL  string `json:"provider_url"`
	HTML         string `json:"html"`
	Width        int64  `json:"width"`
	Height       int64  `json:"height"`
}

// GetFavourites return the favorite list of the current user.
func (c *Client) GetFavourites(ctx context.Context, pg *Pagination) ([]*Status, error) {
	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/favourites", nil, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetStatus return status specified by id.
func (c *Client) GetStatus(ctx context.Context, id string) (*Status, error) {
	var status Status
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/statuses/%s", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// GetStatusContext return status specified by id.
func (c *Client) GetStatusContext(ctx context.Context, id string) (*Context, error) {
	var context Context
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/statuses/%s/context", id), nil, &context, nil)
	if err != nil {
		return nil, err
	}
	return &context, nil
}

// GetStatusCard return status specified by id.
func (c *Client) GetStatusCard(ctx context.Context, id string) (*Card, error) {
	var card Card
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/statuses/%s/card", id), nil, &card, nil)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// GetRebloggedBy returns the account list of the user who reblogged the toot of id.
func (c *Client) GetRebloggedBy(ctx context.Context, id string, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/statuses/%s/reblogged_by", id), nil, &accounts, pg)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetFavouritedBy returns the account list of the user who liked the toot of id.
func (c *Client) GetFavouritedBy(ctx context.Context, id string, pg *Pagination) ([]*Account, error) {
	var accounts []*Account
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/statuses/%s/favourited_by", id), nil, &accounts, pg)
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// Reblog is reblog the toot of id and return status of reblog.
func (c *Client) Reblog(ctx context.Context, id string) (*Status, error) {
	var status Status
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/reblog", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Unreblog is unreblog the toot of id and return status of the original toot.
func (c *Client) Unreblog(ctx context.Context, id string) (*Status, error) {
	var status Status
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/unreblog", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Favourite is favourite the toot of id and return status of the favourite toot.
func (c *Client) Favourite(ctx context.Context, id string) (*Status, error) {
	var status Status
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/favourite", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Unfavourite is unfavourite the toot of id and return status of the unfavourite toot.
func (c *Client) Unfavourite(ctx context.Context, id string) (*Status, error) {
	var status Status
	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/unfavourite", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// GetTimelineHome return statuses from home timeline.
func (c *Client) GetTimelineHome(ctx context.Context, pg *Pagination) ([]*Status, error) {
	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/timelines/home", nil, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetTimelinePublic return statuses from public timeline.
func (c *Client) GetTimelinePublic(ctx context.Context, isLocal bool, pg *Pagination) ([]*Status, error) {
	params := url.Values{}
	if isLocal {
		params.Set("local", "true")
	}

	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/timelines/public", params, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetTimelineHashtag return statuses from tagged timeline.
func (c *Client) GetTimelineHashtag(ctx context.Context, tag string, isLocal bool, pg *Pagination) ([]*Status, error) {
	params := url.Values{}
	if isLocal {
		params.Set("local", "t")
	}

	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/timelines/tag/%s", url.PathEscape(tag)), params, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetTimelineList return statuses from a list timeline.
func (c *Client) GetTimelineList(ctx context.Context, id string, pg *Pagination) ([]*Status, error) {
	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, fmt.Sprintf("/api/v1/timelines/list/%s", url.PathEscape(string(id))), nil, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetTimelineMedia return statuses from media timeline.
// NOTE: This is an experimental feature of pawoo.net.
func (c *Client) GetTimelineMedia(ctx context.Context, isLocal bool, pg *Pagination) ([]*Status, error) {
	params := url.Values{}
	params.Set("media", "t")
	if isLocal {
		params.Set("local", "t")
	}

	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/timelines/public", params, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// PostStatus post the toot.
func (c *Client) PostStatus(ctx context.Context, toot *Toot) (*Status, error) {
	params := url.Values{}
	params.Set("status", toot.Status)
	if toot.InReplyToID != "" {
		params.Set("in_reply_to_id", string(toot.InReplyToID))
	}
	if toot.MediaIDs != nil {
		for _, media := range toot.MediaIDs {
			params.Add("media_ids[]", string(media))
		}
	}
	if toot.Visibility != "" {
		params.Set("visibility", fmt.Sprint(toot.Visibility))
	}
	if toot.Sensitive {
		params.Set("sensitive", "true")
	}
	if toot.SpoilerText != "" {
		params.Set("spoiler_text", toot.SpoilerText)
	}
	if toot.ContentType != "" {
		params.Set("content_type", toot.ContentType)
	}

	var status Status
	err := c.doAPI(ctx, http.MethodPost, "/api/v1/statuses", params, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// DeleteStatus delete the toot.
func (c *Client) DeleteStatus(ctx context.Context, id string) error {
	return c.doAPI(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/statuses/%s", id), nil, nil, nil)
}

// Search search content with query.
func (c *Client) Search(ctx context.Context, q string, qType string, limit int, resolve bool, offset int, accountID string) (*Results, error) {
	var results Results
	params := url.Values{}
	params.Set("q", q)
	params.Set("type", qType)
	params.Set("limit", fmt.Sprint(limit))
	params.Set("resolve", fmt.Sprint(resolve))
	params.Set("offset", fmt.Sprint(offset))
	if len(accountID) > 0 {
		params.Set("account_id", accountID)
	}
	err := c.doAPI(ctx, http.MethodGet, "/api/v2/search", params, &results, nil)
	if err != nil {
		return nil, err
	}
	return &results, nil
}

// UploadMedia upload a media attachment from a file.
func (c *Client) UploadMedia(ctx context.Context, file string) (*Attachment, error) {
	var attachment Attachment
	err := c.doAPI(ctx, http.MethodPost, "/api/v1/media", file, &attachment, nil)
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// UploadMediaFromReader uploads a media attachment from a io.Reader.
func (c *Client) UploadMediaFromReader(ctx context.Context, reader io.Reader) (*Attachment, error) {
	var attachment Attachment
	err := c.doAPI(ctx, http.MethodPost, "/api/v1/media", reader, &attachment, nil)
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// UploadMediaFromReader uploads a media attachment from a io.Reader.
func (c *Client) UploadMediaFromMultipartFileHeader(ctx context.Context, fh *multipart.FileHeader) (*Attachment, error) {
	var attachment Attachment
	err := c.doAPI(ctx, http.MethodPost, "/api/v1/media", fh, &attachment, nil)
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// GetTimelineDirect return statuses from direct timeline.
func (c *Client) GetTimelineDirect(ctx context.Context, pg *Pagination) ([]*Status, error) {
	params := url.Values{}

	var statuses []*Status
	err := c.doAPI(ctx, http.MethodGet, "/api/v1/timelines/direct", params, &statuses, pg)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// MuteConversation mutes status specified by id.
func (c *Client) MuteConversation(ctx context.Context, id string) (*Status, error) {
	var status Status

	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/mute", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// UnmuteConversation unmutes status specified by id.
func (c *Client) UnmuteConversation(ctx context.Context, id string) (*Status, error) {
	var status Status

	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/unmute", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Bookmark bookmarks status specified by id.
func (c *Client) Bookmark(ctx context.Context, id string) (*Status, error) {
	var status Status

	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/bookmark", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// Unbookmark bookmarks status specified by id.
func (c *Client) Unbookmark(ctx context.Context, id string) (*Status, error) {
	var status Status

	err := c.doAPI(ctx, http.MethodPost, fmt.Sprintf("/api/v1/statuses/%s/unbookmark", id), nil, &status, nil)
	if err != nil {
		return nil, err
	}
	return &status, nil
}
