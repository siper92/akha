package task

const (
	IDEA_SECTION_KEY      = "idea"
	QUESTIONS_SECTION_KEY = "questions"
	ANSWERS_SECTION_KEY   = "answers"
	RATE_SECTION_KEY      = "rate"
)

type FlowStep interface {
	Name() string
	Content() []TextSection
}

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

var _ SectionTitle = (*MDTitle)(nil)

type MDTitle string

func (t MDTitle) ID() string {
	return string(t)
}

func (t MDTitle) String() string {
	return string(t)
}

var _ SectionContent = (*Content)(nil)

type Content string

func (c Content) String() string {
	return string(c)
}

var _ TextSection = (*MDTextSection)(nil)

type MDTextSection struct {
	title   SectionTitle
	content SectionContent
	raw     string
}

func (M MDTextSection) Title() SectionTitle {
	return M.title
}

func (M MDTextSection) Content() SectionContent {
	return M.content
}

func (M MDTextSection) String() string {
	return M.raw
}
