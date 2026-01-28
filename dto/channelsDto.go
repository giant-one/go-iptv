package dto

// ChannelDto 频道数据结构
type ChannelDto struct {
	SrcList string `json:"src_list"`
	Ku9     string `json:"ku9"`
}

// OrderedGenreChannelDto 有序的分组频道数据结构
type OrderedGenreChannelDto struct {
	GenreName string
	ChannelDto
}
