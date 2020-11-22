package gotcha

import (
	"strings"
)

type StructField struct {
	Name string
	Type string
	Tags []string
}

func (f []StructField) ToGoCode(structName string) {
	items := []string{}
	for _, field := range f {
		tags := strings.TrimSpace(strings.Join(field.Tags, " "))
		if tags != "" {
			tags = "`" + tags + "`"
		}
		item := fmt.Sprintf("%s    %s    %s", field.Name, field.Type, tags)
		items = append(items, item)
	}

	return fmt.Sprintf("type %s struct {\n%s\n}\n", name, strings.Join(items, "\n"))
}
