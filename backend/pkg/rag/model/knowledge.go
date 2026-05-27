package model

type KnowledgeDoc struct {
	BaseModel
	Title    string `gorm:"size:256;not null" json:"title"`
	Content  string `gorm:"type:text;not null" json:"content"`
	Source   string `gorm:"size:512" json:"source"`
	ChunkIdx int    `gorm:"default:0" json:"chunk_idx"`
}

type KnowledgeQA struct {
	BaseModel
	Question string `gorm:"type:text;not null" json:"question"`
	Answer   string `gorm:"type:text;not null" json:"answer"`
	Category string `gorm:"size:128" json:"category"`
}
