package core

import (
	"fmt"

	yaml "go.yaml.in/yaml/v3"
)

type OrderedVar struct {
	Key   string
	Value string
}

type OrderedVars []OrderedVar

func (ov *OrderedVars) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == 0 {
		*ov = nil
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected YAML mapping node, got %v", node.Kind)
	}

	result := make(OrderedVars, 0, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		key := keyNode.Value

		value, err := scalarToString(valueNode)
		if err != nil {
			return fmt.Errorf("variable %q: %w", key, err)
		}

		result = append(result, OrderedVar{
			Key:   key,
			Value: value,
		})
	}

	*ov = result
	return nil
}

func scalarToString(node *yaml.Node) (string, error) {
	if node.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("expected scalar value, got %v", node.Kind)
	}
	if node.ShortTag() == "!!null" {
		return "", nil
	}
	// node.Value already holds the textual form for !!str/!!int/!!float/!!bool
	return node.Value, nil
}

func (ov OrderedVars) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, item := range ov {
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: item.Key},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: item.Value},
		)
	}

	return node, nil
}
