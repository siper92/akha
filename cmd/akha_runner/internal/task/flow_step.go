package task

import "regexp"

var _ FlowStep = (*TaskFlowStep)(nil)

type TaskFlowStep struct {
	name    string
	content []TextSection
}

func (t TaskFlowStep) Name() string {
	return t.name
}

func (t TaskFlowStep) Content() []TextSection {
	return t.content
}

func ToFlowStep(step []Step) ([]FlowStep, error) {
	var result []FlowStep
	for _, textSection := range step {
		c := textSection.Content()

		idFromContent := getIDFromContent(c)
		if idFromContent == "" {
			idFromContent = textSection.Name()
		}

		parsedDoc, err := ParseMDStepString(c)
		if err != nil {
			return nil, err
		}

		parsedC := []TextSection{}
		for _, section := range parsedDoc.Sections {
			parsedC = append(parsedC, &MDTextSection{
				title:   MDTitle(section.Name),
				content: Content(section.Body),
				raw:     section.Body,
			})
		}

		result = append(result, TaskFlowStep{
			name:    idFromContent,
			content: parsedC,
		})
	}

	return result, nil
}

func parseContent(content string) ([]TextSection, error) {
	return nil, nil
}

func getIDFromContent(content string) string {
	// regex read: line from # @id: to new line
	regex, err := regexp.Compile("# @id: (.*)\n")
	if err != nil {
		return ""
	}

	matches := regex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}
