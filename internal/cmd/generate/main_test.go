package main

import "testing"

func TestParameterVariableAvoidsGoKeywords(t *testing.T) {
	t.Parallel()
	if got := parameterVariable("type"); got != "typeValue" {
		t.Fatalf("parameterVariable(type)=%q", got)
	}
	if got := parameterVariable("object_id"); got != "objectID" {
		t.Fatalf("parameterVariable(object_id)=%q", got)
	}
	if got := parameterVariable("id"); got != "id" {
		t.Fatalf("parameterVariable(id)=%q", got)
	}
}

func TestPathValueExpressionUsesPrimitiveEncoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		parameter parameter
		want      string
	}{
		{parameter: parameter{Name: "type", Schema: schema{"type": "string"}}, want: "string(typeValue)"},
		{parameter: parameter{Name: "version", Schema: schema{"type": "integer"}}, want: "strconv.FormatInt(int64(version), 10)"},
		{parameter: parameter{Name: "enabled", Schema: schema{"type": "boolean"}}, want: "strconv.FormatBool(bool(enabled))"},
	}
	for _, test := range tests {
		got, err := pathValueExpression(test.parameter)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Fatalf("pathValueExpression(%s)=%q, want %q", test.parameter.Name, got, test.want)
		}
	}
}

func TestFlattenedPropertiesAcceptsIdenticalAllOfProperty(t *testing.T) {
	t.Parallel()
	value := schema{"allOf": []any{
		map[string]any{"type": "object", "properties": map[string]any{"locale": map[string]any{"type": "string"}}},
		map[string]any{"type": "object", "properties": map[string]any{"locale": map[string]any{"type": "string"}}},
	}}
	properties, _, err := flattenedProperties(value)
	if err != nil {
		t.Fatal(err)
	}
	if len(properties) != 1 {
		t.Fatalf("properties=%d, want 1", len(properties))
	}
}

func TestFlattenedPropertiesRejectsConflictingAllOfProperty(t *testing.T) {
	t.Parallel()
	value := schema{"allOf": []any{
		map[string]any{"type": "object", "properties": map[string]any{"locale": map[string]any{"type": "string"}}},
		map[string]any{"type": "object", "properties": map[string]any{"locale": map[string]any{"type": "integer"}}},
	}}
	if _, _, err := flattenedProperties(value); err == nil {
		t.Fatal("expected conflicting allOf property to fail")
	}
}
