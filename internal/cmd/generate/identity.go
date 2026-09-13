package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"strings"
)

// Identity has hand-written protocol orchestration; its payloads and operation
// catalog still come from the same canonical contract as other packages.
func (generator *generator) generateIdentity() ([]byte, error) {
	operations, err := generator.operations()
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	fmt.Fprintln(&output, "// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.")
	fmt.Fprintln(&output, "// Source: contracts/openapi/v1/identity.yaml\n\npackage identity")
	fmt.Fprintln(&output, "import (\"net/http\"; \"github.com/aetherplatform/aether-go/internal/transport\")")
	for _, name := range sortedKeys(generator.document.Components.Schemas) {
		if !strings.HasPrefix(name, "Passwordless") {
			continue
		}
		value := generator.document.Components.Schemas[name]
		if variants, ok := value["oneOf"].([]any); ok {
			fmt.Fprintf(&output, "type %s interface { is%s() }\n", name, name)
			for _, raw := range variants {
				variant, _ := raw.(map[string]any)
				typeName := generator.matchingName(variant)
				if typeName == "" {
					return nil, fmt.Errorf("unnamed Identity outcome for %s", name)
				}
				fmt.Fprintf(&output, "func (*%s) is%s() {}\n", typeName, name)
			}
		} else if err := generator.emitNamedSchema(&output, goName(name), value); err != nil {
			return nil, err
		}
	}
	for _, name := range generator.sortedInlineEnumNames() {
		for value, enumName := range generator.inlineEnums {
			if enumName == name {
				var enumSchema schema
				if err := json.Unmarshal([]byte(value), &enumSchema); err != nil {
					return nil, err
				}
				generator.emitEnum(&output, name, enumSchema)
				break
			}
		}
	}
	fmt.Fprintln(&output, "var (")
	for _, op := range operations {
		fmt.Fprintf(&output, "%sOperation = transport.Operation{Name: %q, Method: %s, Path: %q, Idempotency: %s, SuccessStatuses: %#v}\n", op.ID, op.ID, methodConstant(op.Method), op.Path, idempotencyConstant(op.Idempotency), op.SuccessStatuses)
	}
	fmt.Fprintln(&output, ")\nvar allOperations = []transport.Operation{")
	for _, op := range operations {
		fmt.Fprintf(&output, "%sOperation,\n", op.ID)
	}
	fmt.Fprintln(&output, "}")
	return format.Source(output.Bytes())
}
