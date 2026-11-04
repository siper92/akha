package proto

type SectionContent interface {
	String() string
}

type SectionTitle interface {
	ID() string
	String() string
}

type TextSection interface {
	Title() SectionTitle
	Content() SectionContent
	String() string // title + content as in doc
}
