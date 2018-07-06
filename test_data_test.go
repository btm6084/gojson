package gojson

import "fmt"

// TestComponentResponse is the stream Component model
type TestComponentResponse struct {
	Count  int                 `json:"count"`
	Frame  int                 `json:"frame"`
	Items  []TestComponentItem `json:"items"`
	Offset int                 `json:"offset"`
	Status string              `json:"status"`
}

// TestComponentItem models the TestComponent Item
type TestComponentItem struct {
	Category          string                    `json:"category"`
	Data              TestComponentItemData     `json:"data"`
	End               string                    `json:"end"`
	ID                string                    `json:"id"`
	Metadata          TestComponentItemMetadata `json:"metadata"`
	Position          int                       `json:"position,omitempty"`
	Start             string                    `json:"start"`
	TestContentSource string                    `json:"content_source"`
	Type              string                    `json:"type"`
}

// TestComponentItemData models the TestComponent Item Data
type TestComponentItemData struct {
	Assets           []TestComponentItemDataAssets `json:"assets"`
	Categories       []string                      `json:"categories"`
	ComponentURI     string                        `json:"componentUri"`
	ComputedPosition string                        `json:"computedPosition"`
	DataType         string                        `json:"data_type"`
	Position         string                        `json:"position"`
	Summary          string                        `json:"summary"`
	Template         string                        `json:"template"`
	Title            ComponentTitle                `json:"title"`
}

// ComponentTitle handles varying title objects
type ComponentTitle struct {
	Value string
	Type  string
}

// UnmarshalJSON unmarshals json for ComponentTitle
func (d *ComponentTitle) UnmarshalJSON(b []byte) error {
	var complex map[string]string
	var simple string
	var err error

	t := GetJSONType(b, 0)
	switch t {
	case "object":
		err = Unmarshal(b, &complex)
		if err != nil {
			return err
		}

		if _, ok := complex["large"]; ok {
			d.Type = "complex"
			d.Value = complex["large"]
			return nil
		}
	case "string":
		err = Unmarshal(b, &simple)
		if err != nil {
			return err
		}

		d.Type = "simple"
		d.Value = simple
		return nil
	}

	return fmt.Errorf("ComponentTitle: No extraction policy for type '%s'", t)
}

// TestComponentItemDataAssets models assets in the component item data
type TestComponentItemDataAssets struct {
	AdProviderString          string         `json:"ad_provider_string"`
	AssetID                   string         `json:"asset_id"`
	Begins                    string         `json:"begins"`
	CallToAction              string         `json:"call_to_action"`
	CallToActionURL           string         `json:"call_to_action_url"`
	Citation                  string         `json:"citation"`
	ComponentURI              string         `json:"component_uri"`
	HashTag                   string         `json:"hash_tag"`
	HashTagURL                string         `json:"hash_tag_url"`
	Icon                      string         `json:"icon"`
	ID                        string         `json:"id"`
	Image                     string         `json:"image"`
	ImageHeight               string         `json:"image_height"`
	ImageWidth                string         `json:"image_width"`
	IntentURI                 string         `json:"intent_uri"`
	OpenNewWindow             string         `json:"open_new_window"`
	Player                    string         `json:"player"`
	PlayerType                string         `json:"player_type"`
	ProviderKey               string         `json:"provider_key"`
	RawURL                    string         `json:"raw_url"`
	ShortTitle                string         `json:"short_title"`
	SmallImage                string         `json:"small_image"`
	SmallImageHeight          string         `json:"small_image_height"`
	SmallImageWidth           string         `json:"small_image_width"`
	Summary                   string         `json:"summary"`
	Title                     ComponentTitle `json:"title"`
	ThumbnailImage            string         `json:"thumbnail_image"`
	ThumbnailImageHeight      string         `json:"thumbnail_image_height"`
	ThumbnailImageWidth       string         `json:"thumbnail_image_width"`
	ThumbnailTitle            string         `json:"thumbnail_title"`
	ThumbnailSmallImage       string         `json:"thumbnail_small_image"`
	ThumbnailSmallImageHeight string         `json:"thumbnail_small_image_height"`
	ThumbnailSmallImageWidth  string         `json:"thumbnail_small_image_width"`
	Vendor                    string         `json:"vendor"`
	Video                     *Video         `json:"video"`
	VideoID                   string         `json:"video_id"`
}

// TestComponentItemMetadata is the metadata for a stream item
type TestComponentItemMetadata struct {
	CeArticleURL   string   `json:"ce_article_url"`
	CeCitation     string   `json:"ce_citation"`
	CeDescription  string   `json:"ce_description"`
	CeHeadline     string   `json:"ce_headline"`
	CeImage        string   `json:"ce_image"`
	CeURLTarget    string   `json:"ce_url_target"`
	Collections    []string `json:"collections"`
	Devices        []string `json:"devices"`
	MetadataSchema string   `json:"metadata_schema"`
	Network        string   `json:"network"`
	Overrides      []string `json:"overrides"`
	Role           string   `json:"role"`
	Schema         string   `json:"schema"`
	Type           string   `json:"type"`
}

// Video describes a shared video asset
type Video struct {
	Ad            VideoAd            `json:"ad"`
	Available     bool               `json:"available"`
	AssetID       string             `json:"asset_id"`
	ClosedCaption VideoClosedCaption `json:"closed_caption"`
	Bitrate       int                `json:"bitrate"`
	Duration      int                `json:"duration"`
	ID            string             `json:"id"`
	Player        string             `json:"player"`
	Tests         map[string]string  `json:"streams"`
	Codec         interface{}        `json:"codec"`
	Content       []string           `json:"content"`
	Filesize      string             `json:"filesize"`
	FormatID      interface{}        `json:"format_id"`
	Framerate     interface{}        `json:"framerate"`
	Language      string             `json:"language"`
	Md5           interface{}        `json:"md5"`
	MediaScheme   string             `json:"media_scheme"`
	Mimetype      string             `json:"mimetype"`
	Role          interface{}        `json:"role"`
	Samplingrate  interface{}        `json:"samplingrate"`
	Signature     interface{}        `json:"signature"`
	Time          interface{}        `json:"time"`
}

// VideoAd describes a video ad
type VideoAd struct {
	Cutlist []VideoAdCutlist `json:"cutlist"`
	Preroll []VideoAdPreroll `json:"preroll"`
}

// VideoAdCutlist describes a cutlist data
type VideoAdCutlist interface{}

// VideoAdPreroll describes preroll data
type VideoAdPreroll struct {
	Provider       string   `json:"provider"`
	ClientID       string   `json:"client_id"`
	Modified       string   `json:"modified"`
	Created        string   `json:"created"`
	Precedence     int      `json:"precedence"`
	ProviderString string   `json:"provider_string"`
	Clients        []string `json:"clients"`
	TargetType     string   `json:"target_type"`
	ID             string   `json:"id"`
	TargetID       string   `json:"target_id"`
}

// VideoClosedCaption describes the closed captions for a video
type VideoClosedCaption struct {
	Created  string `json:"created"`
	File     string `json:"file"`
	ID       string `json:"id"`
	Language string `json:"language"`
	Modified string `json:"modified"`
	VideoID  string `json:"video_id"`
}

// Image defines images in the response
type Image struct {
	Active       interface{}   `json:"active"`
	Artists      []interface{} `json:"artists"`
	Attribution  interface{}   `json:"attribution"`
	Caption      string        `json:"caption"`
	Created      string        `json:"created"`
	Description  string        `json:"description"`
	Descriptions []string      `json:"descriptions"`
	FeedKey      string        `json:"feed_key"`
	Filesize     int           `json:"filesize"`
	GroupID      string        `json:"group_id"`
	Height       int           `json:"height"`
	ID           string        `json:"id"`
	Language     interface{}   `json:"language"`
	Link         string        `json:"link"`
	Md5          interface{}   `json:"md5"`
	Mimetype     string        `json:"mimetype"`
	Modified     string        `json:"modified"`
	Orientation  interface{}   `json:"orientation"`
	OriginalID   interface{}   `json:"original_id"`
	ParentID     string        `json:"parent_id"`
	Preferred    int           `json:"preferred"`
	Processed    interface{}   `json:"processed"`
	Role         string        `json:"role"`
	Signature    string        `json:"signature"`
	Time         interface{}   `json:"time"`
	URL          string        `json:"url"`
	Width        int           `json:"width"`
}

var (
	tdEmptyString = []byte(`""`)
	tdString      = []byte(`"some string"`)
	tdInt         = []byte(`17`)
	tdBool        = []byte(`true`)
	tdNull        = []byte(`null`)
	tdFloat       = []byte(`22.83`)
	tdStringSlice = []byte(`[ "a", "b", "c", "d", "e", "t" ]`)
	tdBoolSlice   = []byte(`[ true, false, true, false ]`)
	tdIntSlice    = []byte(`[ -1, 0, 1, 2, 3, 4 ]`)
	tdFloatSlice  = []byte(`[ -1.1, 0.0, 1.1, 2.2, 3.3 ]`)
	tdObject      = []byte(`{ "a": "b", "c": "d" }`)
	tdObjects     = []byte(`[ { "e": "f", "g": "h" }, { "i": "j", "k": "l" }, { "m": "n", "o": "t" } ]`)
	tdComplex     = []byte(`[ "a", 2, null, false, 2.2, { "c": "d", "empty_string": "" }, [ "s" ] ]`)

	readerTestData = []byte(`{
		"empty_string": "",
		"string": "some string",
		"int": 17,
		"bool": true,
		"null": null,
		"float": 22.83,
		"string_slice": [ "a", "b", "c", "d", "e", "t" ],
		"bool_slice": [ true, false, true, false ],
		"int_slice": [ -1, 0, 1, 2, 3, 4 ],
		"float_slice": [ -1.1, 0.0, 1.1, 2.2, 3.3 ],
		"object": { "a": "b", "c": "d" },
		"objects": [
			{ "e": "f", "g": "h" },
			{ "i": "j", "k": "l" },
			{ "m": "n", "o": "t" }
		],
		"complex": [ "a", 2, null, false, 2.2, { "c": "d", "empty_string": "" }, [ "s" ] ]
	}`)

	readerTestDataKeys = []string{
		"empty_string",
		"string",
		"int",
		"bool",
		"null",
		"float",
		"string_slice",
		"bool_slice",
		"int_slice",
		"float_slice",
		"object",
		"objects",
		"complex",
	}
)

var largeJSONTestBlob = `{
	"count": 19,
	"status": "OK",
	"frame": 0,
	"offset": 0,
	"items": [
		{
			"id": "id 0",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 0",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "slideshow",
						"thumbnail_image": "",
						"thumbnail_small_image_height": "",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "",
						"citation": "citation 0",
						"hash_tag_url": "",
						"hash_tag": "Sports",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/0",
						"component_uri": "",
						"image_width": "770",
						"player": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/0",
						"short_title": "",
						"raw_url": "http://www.example.com/raw_url/0",
						"thumbnail_image_height": "",
						"thumbnail_small_image": "",
						"intent_uri": "",
						"thumbnail_title": "",
						"thumbnail_image_width": "",
						"asset_id": "",
						"title": "title 0",
						"vendor": ""
					}
				],
				"template": "default",
				"computedPosition": "1"
			},
			"metadata": {
				"schema": "schema 0",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 0",
			"content_source": "content_source 0"
		},
		{
			"id": "id 1",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 1",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"thumbnail_image": "http://www.example.com/thumbnail_image/1",
						"thumbnail_small_image_height": "516",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 1",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "768",
						"citation": "citation 1",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/1",
						"player": "",
						"component_uri": "",
						"image_width": "770",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/1",
						"short_title": "short title 1",
						"raw_url": "http://www.example.com/raw_url/1",
						"thumbnail_image_height": "516",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/1",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 1",
						"thumbnail_image_width": "768",
						"asset_id": "asset id 1",
						"title": "title 1",
						"vendor": "",
						"id": "id 1",
						"begins": "2018-02-09 19:41:33",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 1"
					}
				],
				"template": "default",
				"computedPosition": "2"
			},
			"metadata": {
				"schema": "schema 1",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides",
					"news_collection"
				]
			},
			"category": "category 1",
			"content_source": "content_source 1"
		},
		{
			"id": "id 2",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 2",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "video",
						"call_to_action": "",
						"summary": "summary 2",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "video_id_2",
						"call_to_action_url": "",
						"citation": "citation 2",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/2",
						"player": "sf",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "ad_provider_string 2",
						"raw_url": "http://www.example.com/raw_url/2",
						"image": "http://www.example.com/image/2",
						"short_title": "short title 2",
						"intent_uri": "",
						"player_type": "sf",
						"asset_id": "asset id 2",
						"title": "title 2",
						"vendor": "vendor 2",
						"id": "id 2",
						"begins": "2018-02-09 19:07:09",
						"video": {
							"player": "sf",
							"id": "id 2",
							"available": true,
							"streams": {
								"mp4": "http://www.example.com/mp4/2",
								"hls": "http://www.example.com/hls/2",
								"dash": "http://www.example.com/dash/2",
								"hds": "http://www.example.com/hds/2"
							},
							"ad": {
								"cutlist": [],
								"preroll": [
									{
										"provider": "provider 2",
										"client_id": "client_id 2",
										"modified": "2017-10-26 19:25:55.561206",
										"created": "2017-10-26 19:25:55.561206",
										"precedence": 0,
										"provider_string": "provider_string 2",
										"clients": [
											"82122303"
										],
										"target_type": "vendor",
										"id": "id 2",
										"target_id": "target_id 2"
									}
								]
							},
							"duration": 86
						},
						"provider_key": "provider_key 2"
					}
				],
				"template": "default",
				"computedPosition": "3"
			},
			"metadata": {
				"schema": "schema 2",
				"overrides": [
					"default"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 2",
			"content_source": "content_source 2"
		},
		{
			"id": "id 3",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 3",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "video",
						"thumbnail_image": "http://www.example.com/thumbnail_image/3",
						"thumbnail_small_image_height": "1080",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 3",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "1920",
						"citation": "citation 3",
						"hash_tag_url": "",
						"hash_tag": "Sports",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/3",
						"image_width": "770",
						"player": "condenast",
						"component_uri": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/3",
						"short_title": "short title 3",
						"raw_url": "http://www.example.com/raw_url/3",
						"thumbnail_image_height": "1080",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/3",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 3",
						"thumbnail_image_width": "1920",
						"asset_id": "asset id 3",
						"title": "title 3",
						"vendor": "vendor 3",
						"id": "id 3",
						"begins": "2018-02-09 13:00:00",
						"video": {
							"player": "condenast",
							"id": "id 3",
							"available": true,
							"ad": {
								"cutlist": [],
								"preroll": []
							},
							"duration": 124
						},
						"provider_key": "provider_key 3"
					}
				],
				"template": "default",
				"computedPosition": "4"
			},
			"metadata": {
				"schema": "schema 3",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 3",
			"content_source": "content_source 3"
		},
		{
			"id": "id 4",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 4",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 4",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 4",
						"hash_tag_url": "",
						"hash_tag": "Entertainment",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/4",
						"player": "",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/4",
						"image": "http://www.example.com/image/4",
						"short_title": "short title 4",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 4",
						"title": "title 4",
						"vendor": "",
						"id": "id 4",
						"begins": "2018-02-09 15:00:00",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 4"
					}
				],
				"template": "default",
				"computedPosition": "5"
			},
			"metadata": {
				"schema": "schema 4",
				"overrides": [
					"default"
				],
				"collections": [
					"component_slides",
					"spaces_entertainment"
				]
			},
			"category": "category 4",
			"content_source": "content_source 4"
		},
		{
			"id": "id 5",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 5",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 5",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 5",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/5",
						"player": "",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/5",
						"image": "http://www.example.com/image/5",
						"short_title": "short title 5",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 5",
						"title": "title 5",
						"vendor": "",
						"id": "id 5",
						"begins": "2018-02-09 16:45:00",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 5"
					}
				],
				"template": "default",
				"computedPosition": "6"
			},
			"metadata": {
				"schema": "schema 5",
				"overrides": [
					"default"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 5",
			"content_source": "content_source 5"
		},
		{
			"id": "id 6",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 6",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 6",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 6",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/6",
						"player": "",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/6",
						"image": "http://www.example.com/image/6",
						"short_title": "short title 6",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 6",
						"title": "title 6",
						"vendor": "",
						"id": "id 6",
						"begins": "2018-02-09 14:13:55",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 6"
					}
				],
				"template": "default",
				"computedPosition": "7"
			},
			"metadata": {
				"schema": "schema 6",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"news_collection",
					"component_slides"
				]
			},
			"category": "category 6",
			"content_source": "content_source 6"
		},
		{
			"id": "id 7",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 7",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 7",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 7",
						"hash_tag_url": "",
						"hash_tag": "Entertainment",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/7",
						"player": "",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/7",
						"image": "http://www.example.com/image/7",
						"short_title": "short title 7",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 7",
						"title": "title 7",
						"vendor": "",
						"id": "id 7",
						"begins": "2018-02-09 18:52:13",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 7"
					}
				],
				"template": "default",
				"computedPosition": "8"
			},
			"metadata": {
				"schema": "schema 7",
				"overrides": [
					"default"
				],
				"collections": [
					"spaces_entertainment",
					"component_slides"
				]
			},
			"category": "category 7",
			"content_source": "content_source 7"
		},
		{
			"id": "id 8",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 8",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 8",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 8",
						"hash_tag_url": "",
						"hash_tag": "Sports",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/8",
						"player": "",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/8",
						"image": "http://www.example.com/image/8",
						"short_title": "short title 8",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 8",
						"title": "title 8",
						"vendor": "",
						"id": "id 8",
						"begins": "2018-02-09 12:13:17",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 8"
					}
				],
				"template": "default",
				"computedPosition": "9"
			},
			"metadata": {
				"schema": "schema 8",
				"overrides": [
					"default"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 8",
			"content_source": "content_source 8"
		},
		{
			"id": "id 9",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 9",
				"assets": [
					{
						"small_image_height": "1414",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 9",
						"image_height": "1414",
						"small_image_width": "2121",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 9",
						"hash_tag_url": "",
						"hash_tag": "Finance",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/9",
						"player": "",
						"image_width": "2121",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/9",
						"image": "http://www.example.com/image/9",
						"short_title": "short title 9",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 9",
						"title": "title 9",
						"vendor": "",
						"id": "id 9",
						"begins": "2018-02-09 13:11:03",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 9"
					}
				],
				"template": "default",
				"computedPosition": "10"
			},
			"metadata": {
				"schema": "schema 9",
				"overrides": [
					"default"
				],
				"collections": [
					"component_slides",
					"spaces_finance"
				]
			},
			"category": "category 9",
			"content_source": "content_source 9"
		},
		{
			"id": "id 10",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 10",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"call_to_action": "",
						"summary": "summary 10",
						"image_height": "252",
						"small_image_width": "640",
						"video_id": "",
						"call_to_action_url": "",
						"citation": "citation 10",
						"hash_tag_url": "",
						"hash_tag": "Entertainment",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/10",
						"player": "",
						"image_width": "770",
						"component_uri": "",
						"ad_provider_string": "",
						"raw_url": "http://www.example.com/raw_url/10",
						"image": "http://www.example.com/image/10",
						"short_title": "short title 10",
						"intent_uri": "",
						"player_type": "",
						"asset_id": "asset id 10",
						"title": "title 10",
						"vendor": "",
						"id": "id 10",
						"begins": "2018-02-08 18:59:51",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 10"
					}
				],
				"template": "default",
				"computedPosition": "11"
			},
			"metadata": {
				"schema": "schema 10",
				"overrides": [
					"default"
				],
				"collections": [
					"spaces_entertainment",
					"component_slides"
				]
			},
			"category": "category 10",
			"content_source": "content_source 10"
		},
		{
			"id": "id 11",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 11",
				"assets": [
					{
						"small_image_height": "360",
						"icon": "video",
						"thumbnail_image": "http://www.example.com/thumbnail_image/11",
						"thumbnail_small_image_height": "360",
						"call_to_action": "",
						"small_image_width": "636",
						"summary": "summary 11",
						"image_height": "209",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "640",
						"citation": "citation 11",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/11",
						"player": "sf",
						"component_uri": "",
						"image_width": "640",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/11",
						"short_title": "short title 11",
						"raw_url": "http://www.example.com/raw_url/11",
						"thumbnail_image_height": "360",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/11",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 11",
						"thumbnail_image_width": "640",
						"asset_id": "asset id 11",
						"title": "title 11",
						"vendor": "vendor 11",
						"id": "id 11",
						"begins": "2018-02-09 15:55:00",
						"video": {
							"player": "sf",
							"id": "id 11",
							"available": true,
							"streams": {
								"mp4": "http://www.example.com/mp4/11",
								"hls": "http://www.example.com/hls/11",
								"dash": "http://www.example.com/dash/11",
								"hds": "http://www.example.com/hds/11"
							},
							"ad": {
								"cutlist": [],
								"preroll": [
									{
										"provider": "provider 11",
										"client_id": "client_id 11",
										"modified": "2017-10-02 17:21:13.252817",
										"created": "2016-08-12 14:31:14.344145",
										"precedence": 0,
										"provider_string": "provider_string 11",
										"clients": [
											"82122303"
										],
										"target_type": "vendor",
										"id": "id 11",
										"target_id": "target_id 11"
									}
								]
							},
							"closed_caption": {
								"created": "2018-02-09 16:28:49.096577",
								"language": "en",
								"video_id": "220197962",
								"file": "http://www.example.com/file/11",
								"id": "id 11",
								"modified": "2018-02-09 16:28:49.096577"
							},
							"duration": 43
						},
						"provider_key": "provider_key 11"
					}
				],
				"template": "default",
				"computedPosition": "12"
			},
			"metadata": {
				"schema": "schema 11",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"spaces_finance",
					"component_slides"
				]
			},
			"category": "category 11",
			"content_source": "content_source 11"
		},
		{
			"id": "id 12",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 12",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"thumbnail_image": "http://www.example.com/thumbnail_image/12",
						"thumbnail_small_image_height": "1066",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 12",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "1600",
						"citation": "citation 12",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/12",
						"component_uri": "",
						"image_width": "770",
						"player": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/12",
						"short_title": "short title 12",
						"raw_url": "http://www.example.com/raw_url/12",
						"thumbnail_image_height": "1066",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/12",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 12",
						"thumbnail_image_width": "1600",
						"asset_id": "asset id 12",
						"title": "title 12",
						"vendor": "",
						"id": "id 12",
						"begins": "2018-02-09 17:01:57",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 12"
					}
				],
				"template": "default",
				"computedPosition": "13"
			},
			"metadata": {
				"schema": "schema 12",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 12",
			"content_source": "content_source 12"
		},
		{
			"id": "id 13",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 13",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "video",
						"thumbnail_image": "http://www.example.com/thumbnail_image/13",
						"thumbnail_small_image_height": "438",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 13",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "780",
						"citation": "citation 13",
						"hash_tag_url": "",
						"hash_tag": "Sports",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/13",
						"player": "newspeople",
						"component_uri": "",
						"image_width": "770",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/13",
						"short_title": "short title 13",
						"raw_url": "http://www.example.com/raw_url/13",
						"thumbnail_image_height": "438",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/13",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 13",
						"thumbnail_image_width": "780",
						"asset_id": "asset id 13",
						"title": "title 13",
						"vendor": "vendor 13",
						"id": "id 13",
						"begins": "2018-02-09 11:28:34",
						"video": {
							"player": "newspeople",
							"id": "id 13",
							"available": true,
							"ad": {
								"cutlist": [],
								"preroll": []
							},
							"duration": 58
						},
						"provider_key": "provider_key 13"
					}
				],
				"template": "default",
				"computedPosition": "14"
			},
			"metadata": {
				"schema": "schema 13",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 13",
			"content_source": "content_source 13"
		},
		{
			"id": "id 14",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 14",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"thumbnail_image": "http://www.example.com/thumbnail_image/14",
						"thumbnail_small_image_height": "1080",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 14",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "1920",
						"citation": "citation 14",
						"hash_tag_url": "",
						"hash_tag": "Sports",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/14",
						"image_width": "770",
						"player": "",
						"component_uri": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/14",
						"short_title": "short title 14",
						"raw_url": "http://www.example.com/raw_url/14",
						"thumbnail_image_height": "1080",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/14",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 14",
						"thumbnail_image_width": "1920",
						"asset_id": "asset id 14",
						"title": "title 14",
						"vendor": "",
						"id": "id 14",
						"begins": "2018-02-09 12:24:38",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 14"
					}
				],
				"template": "default",
				"computedPosition": "15"
			},
			"metadata": {
				"schema": "schema 14",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"spaces_sports",
					"component_slides"
				]
			},
			"category": "category 14",
			"content_source": "content_source 14"
		},
		{
			"id": "id 15",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 15",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"thumbnail_image": "http://www.example.com/thumbnail_image/15",
						"thumbnail_small_image_height": "1255",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 15",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "2231",
						"citation": "citation 15",
						"hash_tag_url": "",
						"hash_tag": "Entertainment",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/15",
						"component_uri": "",
						"image_width": "770",
						"player": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/15",
						"short_title": "short title 15",
						"raw_url": "http://www.example.com/raw_url/15",
						"thumbnail_image_height": "1255",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/15",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 15",
						"thumbnail_image_width": "2231",
						"asset_id": "asset id 15",
						"title": "title 15",
						"vendor": "",
						"id": "id 15",
						"begins": "2018-02-09 16:11:46",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 15"
					}
				],
				"template": "default",
				"computedPosition": "16"
			},
			"metadata": {
				"schema": "schema 15",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides",
					"spaces_entertainment"
				]
			},
			"category": "category 15",
			"content_source": "content_source 15"
		},
		{
			"id": "id 16",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 16",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"thumbnail_image": "http://www.example.com/thumbnail_image/16",
						"thumbnail_small_image_height": "720",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 16",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "1280",
						"citation": "citation 16",
						"hash_tag_url": "",
						"hash_tag": "Entertainment",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/16",
						"component_uri": "",
						"image_width": "770",
						"player": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/16",
						"short_title": "short title 16",
						"raw_url": "http://www.example.com/raw_url/16",
						"thumbnail_image_height": "720",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/16",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 16",
						"thumbnail_image_width": "1280",
						"asset_id": "asset id 16",
						"title": "title 16",
						"vendor": "",
						"id": "id 16",
						"begins": "2018-02-09 15:15:00",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": []
							}
						},
						"provider_key": "provider_key 16"
					}
				],
				"template": "default",
				"computedPosition": "17"
			},
			"metadata": {
				"schema": "schema 16",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 16",
			"content_source": "content_source 16"
		},
		{
			"id": "id 17",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 17",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "video",
						"thumbnail_image": "http://www.example.com/thumbnail_image/17",
						"thumbnail_small_image_height": "438",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 17",
						"image_height": "252",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "780",
						"citation": "citation 17",
						"hash_tag_url": "",
						"hash_tag": "Sports",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/17",
						"component_uri": "",
						"image_width": "770",
						"player": "newspeople",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/17",
						"short_title": "short title 17",
						"raw_url": "http://www.example.com/raw_url/17",
						"thumbnail_image_height": "438",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/17",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 17",
						"thumbnail_image_width": "780",
						"asset_id": "asset id 17",
						"title": "title 17",
						"vendor": "vendor 17",
						"id": "id 17",
						"begins": "2018-02-09 13:28:18",
						"video": {
							"player": "newspeople",
							"id": "id 17",
							"available": true,
							"ad": {
								"cutlist": [],
								"preroll": []
							},
							"duration": 54
						},
						"provider_key": "provider_key 17"
					}
				],
				"template": "default",
				"computedPosition": "18",
				"categories": [
					"sports"
				]
			},
			"metadata": {
				"schema": "schema 17",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"spaces_sports",
					"component_slides"
				]
			},
			"category": "category 17",
			"content_source": "content_source 17"
		},
		{
			"id": "id 18",
			"type": "default",
			"start": "0",
			"end": "0",
			"data": {
				"data_type": "data_type 18",
				"assets": [
					{
						"small_image_height": "362",
						"icon": "none",
						"thumbnail_image": "http://www.example.com/thumbnail_image/18",
						"thumbnail_small_image_height": "362",
						"call_to_action": "",
						"small_image_width": "640",
						"summary": "summary 18",
						"image_height": "770",
						"call_to_action_url": "",
						"thumbnail_small_image_width": "640",
						"citation": "citation 18",
						"hash_tag_url": "",
						"hash_tag": "News",
						"open_new_window": "no",
						"small_image": "http://www.example.com/small_image/18",
						"component_uri": "",
						"image_width": "252",
						"player": "",
						"ad_provider_string": "",
						"image": "http://www.example.com/image/18",
						"short_title": "short title 18",
						"raw_url": "http://www.example.com/raw_url/18",
						"thumbnail_image_height": "640",
						"thumbnail_small_image": "http://www.example.com/thumbnail_small_image/18",
						"intent_uri": "",
						"thumbnail_title": "thumbnail_title 18",
						"thumbnail_image_width": "362",
						"asset_id": "asset id 18",
						"title": "title 18",
						"vendor": "",
						"id": "id 18",
						"begins": "2018-02-08 22:27:50",
						"video": {
							"available": false,
							"ad": {
								"cutlist": [],
								"preroll": {}
							}
						},
						"provider_key": "provider_key 18"
					}
				],
				"template": "default",
				"computedPosition": "19"
			},
			"metadata": {
				"schema": "schema 18",
				"overrides": [
					"client_identifier"
				],
				"collections": [
					"component_slides"
				]
			},
			"category": "category 18",
			"content_source": "content_source 18"
		}
	]
}`
