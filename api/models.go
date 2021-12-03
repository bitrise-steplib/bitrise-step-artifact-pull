package api

import (
	"encoding/json"

	"github.com/markbates/pop/nulls"
)

// ArtifactListResponseModel ...
type ArtifactListResponseModel struct {
	Data []ArtifactListElementResponseModel `json:"data"`
	// pagination
	Paging PagingModel `json:"paging"`
}

// ListBuildArtifactsResponse ...
type ListBuildArtifactsResponse struct {
	Data   []ArtifactListElementResponseModel `json:"data"`
	Paging PagingModel                        `json:"paging"`
}

// ShowBuildArtifact ...
type ShowBuildArtifactResponse struct {
	Data ArtifactResponseItemModel `json:"data"`
}

// ArtifactResponseItemModel ...
type ArtifactResponseItemModel struct {
	Title                nulls.String    `json:"title"`
	ArtifactType         nulls.String    `json:"artifact_type"`
	ArtifactMeta         json.RawMessage `json:"artifact_meta"`
	DownloadPath         *string         `json:"expiring_download_url"`
	IsPublicPageEnabled  bool            `json:"is_public_page_enabled"`
	Slug                 string          `json:"slug"`
	PublicInstallPageURL string          `json:"public_install_page_url"`
	FileSizeBytes        *int64          `json:"file_size_bytes"`
}

// ArtifactListElementResponseModel ...
type ArtifactListElementResponseModel struct {
	Title               nulls.String    `json:"title"`
	ArtifactType        nulls.String    `json:"artifact_type"`
	ArtifactMeta        json.RawMessage `json:"artifact_meta"`
	IsPublicPageEnabled bool            `json:"is_public_page_enabled"`
	Slug                string          `json:"slug"`
	FileSizeBytes       *int64          `json:"file_size_bytes"`
}

type PagingModel struct {
	// TotalItemCount - total item count, through "all pages"
	TotalItemCount int64 `json:"total_item_count"`
	// PageItemLimit - per-page item count. A given page might include
	// less items if there's not enough items. This value is the "max item count per page".
	PageItemLimit uint `json:"page_item_limit"`
	// Next is the "anchor" for pagination. This value should be passed to the same endpoint
	// to get the next page. Empty/not included if there's no "next" page.
	// Stop paging when there's no "Next" item in the response!
	Next string `json:"next,omitempty"`
}

type GetBuildArtifactsResponse struct {
}
