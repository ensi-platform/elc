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

		var value string
		if err := valueNode.Decode(&value); err != nil {
			return err
		}

		result = append(result, OrderedVar{
			Key:   keyNode.Value,
			Value: value,
		})
	}

	*ov = result
	return nil
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
