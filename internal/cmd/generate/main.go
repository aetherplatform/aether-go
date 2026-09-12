package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type schema map[string]any

type document struct {
	Components struct {
		Schemas map[string]schema `json:"schemas"`
	} `json:"components"`
	Paths map[string]map[string]any `json:"paths"`
}

type generator struct {
	platform       string
	document       document
	namedBySchema  map[string]string
	inlineEnums    map[string]string
	operationNames map[string]struct{}
}

type operation struct {
	ID              string
	Method          string
	Path            string
	Idempotency     string
	SuccessStatuses []int
	PathParameters  []parameter
	QueryParameters []parameter
	Request         schema
	Response        schema
}

type parameter struct {
	Name     string
	Location string
	Required bool
	Schema   schema
}

func main() {
	if len(os.Args) != 4 {
		fatalf("usage: generate PLATFORM INPUT.json OUTPUT.go")
	}
	platform, inputPath, outputPath := os.Args[1], os.Args[2], os.Args[3]
	data, err := os.ReadFile(inputPath)
	if err != nil {
		fatalf("read contract: %v", err)
	}
	var contract document
	if err := json.Unmarshal(data, &contract); err != nil {
		fatalf("parse contract: %v", err)
	}
	generated, err := newGenerator(platform, contract).generate()
	if err != nil {
		fatalf("generate: %v", err)
	}
	if err := os.WriteFile(outputPath, generated, 0o644); err != nil {
		fatalf("write output: %v", err)
	}
}

func newGenerator(platform string, contract document) *generator {
	named := make(map[string]string)
	names := sortedKeys(contract.Components.Schemas)
	for _, name := range names {
		named[canonical(contract.Components.Schemas[name])] = goName(name)
	}
	generator := &generator{
		platform:       platform,
		document:       contract,
		namedBySchema:  named,
		inlineEnums:    make(map[string]string),
		operationNames: make(map[string]struct{}),
	}
	for _, name := range names {
		generator.collectInlineEnums(contract.Components.Schemas[name], name)
	}
	return generator
}

func (generator *generator) generate() ([]byte, error) {
	operations, err := generator.operations()
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	fmt.Fprintln(&output, "// Code generated from the canonical Aether OpenAPI contract. DO NOT EDIT.")
	fmt.Fprintf(&output, "// Source: contracts/openapi/v1/%s.yaml\n\n", generator.platform)
	fmt.Fprintf(&output, "package %s\n\n", generator.platform)
	fmt.Fprintln(&output, "import (")
	fmt.Fprintln(&output, "\t\"context\"")
	fmt.Fprintln(&output, "\t\"net/http\"")
	fmt.Fprintln(&output, "\t\"net/url\"")
	fmt.Fprintln(&output, "\t\"strconv\"")
	for _, operation := range operations {
		if operation.Idempotency == "request_field" {
			fmt.Fprintln(&output, "\t\"strings\"")
			break
		}
	}
	fmt.Fprintln(&output, "\t\"time\"")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "\t\"github.com/aetherplatform/aether-go\"")
	fmt.Fprintln(&output, "\t\"github.com/aetherplatform/aether-go/internal/transport\"")
	fmt.Fprintln(&output, ")")

	for _, name := range sortedKeys(generator.document.Components.Schemas) {
		if err := generator.emitNamedSchema(&output, goName(name), generator.document.Components.Schemas[name]); err != nil {
			return nil, fmt.Errorf("schema %s: %w", name, err)
		}
	}
	for _, name := range generator.sortedInlineEnumNames() {
		var enumSchema schema
		for value, enumName := range generator.inlineEnums {
			if enumName == name {
				if err := json.Unmarshal([]byte(value), &enumSchema); err != nil {
					return nil, err
				}
				break
			}
		}
		generator.emitEnum(&output, name, enumSchema)
	}

	for _, operation := range operations {
		generator.emitParams(&output, operation)
	}
	fmt.Fprintln(&output, "var (")
	for _, operation := range operations {
		fmt.Fprintf(&output, "\t%sOperation = transport.Operation{Name: %q, Method: %s, Path: %q, Idempotency: %s, SuccessStatuses: %#v",
			operation.ID, operation.ID, methodConstant(operation.Method), operation.Path, idempotencyConstant(operation.Idempotency), operation.SuccessStatuses)
		if operation.Idempotency == "request_field" {
			if err := generator.emitRequestRetrySafe(&output, operation); err != nil {
				return nil, fmt.Errorf("operation %s: %w", operation.ID, err)
			}
		}
		fmt.Fprintln(&output, "}")
	}
	fmt.Fprintln(&output, ")")
	fmt.Fprintln(&output, "\nvar allOperations = []transport.Operation{")
	for _, operation := range operations {
		fmt.Fprintf(&output, "\t%sOperation,\n", operation.ID)
	}
	fmt.Fprintln(&output, "}")

	for _, operation := range operations {
		if err := generator.emitMethod(&output, operation); err != nil {
			return nil, fmt.Errorf("operation %s: %w", operation.ID, err)
		}
	}
	fmt.Fprintln(&output, "func setString(query url.Values, name string, value *string) { if value != nil { query.Set(name, *value) } }")
	fmt.Fprintln(&output, "func setInteger(query url.Values, name string, value *int64) { if value != nil { query.Set(name, strconv.FormatInt(*value, 10)) } }")
	fmt.Fprintln(&output, "func setTime(query url.Values, name string, value *time.Time) { if value != nil { query.Set(name, value.Format(time.RFC3339)) } }")

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w\n%s", err, output.String())
	}
	return formatted, nil
}

func (generator *generator) emitRequestRetrySafe(output *bytes.Buffer, operation operation) error {
	properties, required, err := flattenedProperties(operation.Request)
	if err != nil {
		return err
	}
	key := properties["idempotency_key"]
	if primitiveType(key) != "string" {
		return fmt.Errorf("request_field requires a string idempotency_key property")
	}
	requestType, _, err := generator.goType(operation.Request, exported(operation.ID)+"Request", "")
	if err != nil {
		return err
	}
	check := "strings.TrimSpace(string(request.IdempotencyKey)) != \"\""
	if _, isRequired := required["idempotency_key"]; !isRequired || isNullable(key) {
		check = "request.IdempotencyKey != nil && strings.TrimSpace(string(*request.IdempotencyKey)) != \"\""
	}
	fmt.Fprintf(output, ", RequestRetrySafe: func(body any) bool { request, ok := body.(%s); return ok && %s }", requestType, check)
	return nil
}

func (generator *generator) emitNamedSchema(output *bytes.Buffer, name string, value schema) error {
	if enumValues(value) != nil {
		generator.emitEnum(output, name, value)
		return nil
	}
	if _, ok := value["const"]; ok {
		fmt.Fprintf(output, "\ntype %s string\n", name)
		return nil
	}
	if isObject(value) || value["allOf"] != nil {
		if value["additionalProperties"] == true && value["properties"] == nil {
			fmt.Fprintf(output, "\ntype %s map[string]any\n", name)
			return nil
		}
		properties, required, err := flattenedProperties(value)
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "\ntype %s struct {\n", name)
		for _, propertyName := range sortedKeys(properties) {
			property := properties[propertyName]
			typeName, nullable, err := generator.goType(property, name+goName(propertyName), name)
			if err != nil {
				return fmt.Errorf("property %s: %w", propertyName, err)
			}
			_, isRequired := required[propertyName]
			if nullable || !isRequired {
				typeName = pointerType(typeName)
			}
			tag := propertyName
			if !isRequired {
				tag += ",omitempty"
			}
			fmt.Fprintf(output, "\t%s %s `json:%q`\n", goName(propertyName), typeName, tag)
		}
		fmt.Fprintln(output, "}")
		return nil
	}
	typeName, _, err := generator.goType(value, name, name)
	if err != nil {
		return err
	}
	fmt.Fprintf(output, "\ntype %s %s\n", name, typeName)
	return nil
}

func (generator *generator) emitEnum(output *bytes.Buffer, name string, value schema) {
	fmt.Fprintf(output, "\ntype %s string\n\nconst (\n", name)
	for _, enumValue := range enumValues(value) {
		fmt.Fprintf(output, "\t%s%s %s = %q\n", name, goName(enumValue), name, enumValue)
	}
	fmt.Fprintln(output, ")")
}

func (generator *generator) emitParams(output *bytes.Buffer, operation operation) {
	if len(operation.QueryParameters) == 0 {
		return
	}
	fmt.Fprintf(output, "\ntype %sParams struct {\n", exported(operation.ID))
	for _, parameter := range operation.QueryParameters {
		typeName, nullable, err := generator.goType(parameter.Schema, exported(operation.ID)+goName(parameter.Name), "")
		if err != nil {
			panic(err)
		}
		if nullable || !parameter.Required {
			typeName = pointerType(typeName)
		}
		fmt.Fprintf(output, "\t%s %s\n", goName(parameter.Name), typeName)
	}
	fmt.Fprintln(output, "}")
}

func (generator *generator) emitMethod(output *bytes.Buffer, operation operation) error {
	name := exported(operation.ID)
	arguments := []string{"ctx context.Context"}
	for _, parameter := range operation.PathParameters {
		typeName, _, err := generator.goType(parameter.Schema, name+goName(parameter.Name), "")
		if err != nil {
			return err
		}
		arguments = append(arguments, parameterVariable(parameter.Name)+" "+typeName)
	}
	requestType := ""
	if operation.Request != nil {
		var err error
		requestType, _, err = generator.goType(operation.Request, name+"Request", "")
		if err != nil {
			return err
		}
		arguments = append(arguments, "request "+requestType)
	}
	if len(operation.QueryParameters) > 0 {
		arguments = append(arguments, "params "+name+"Params")
	}
	arguments = append(arguments, "options ...aether.RequestOption")
	responseType, _, err := generator.goType(operation.Response, name+"Response", "")
	if err != nil {
		return err
	}
	returnType := "*" + responseType
	if strings.HasPrefix(responseType, "[]") {
		returnType = responseType
	}
	fmt.Fprintf(output, "\nfunc (client *Client) %s(%s) (%s, error) {\n", name, strings.Join(arguments, ", "), returnType)
	if len(operation.QueryParameters) > 0 {
		fmt.Fprintln(output, "\tquery := make(url.Values)")
		for _, parameter := range operation.QueryParameters {
			generator.emitQuery(output, parameter)
		}
	} else {
		fmt.Fprintln(output, "\tvar query url.Values")
	}
	if len(operation.PathParameters) > 0 {
		fmt.Fprintln(output, "\tpath := map[string]string{")
		for _, parameter := range operation.PathParameters {
			expression, err := pathValueExpression(parameter)
			if err != nil {
				return err
			}
			fmt.Fprintf(output, "\t\t%q: %s,\n", parameter.Name, expression)
		}
		fmt.Fprintln(output, "\t}")
	} else {
		fmt.Fprintln(output, "\tvar path map[string]string")
	}
	bodyExpression := "nil"
	if requestType != "" {
		bodyExpression = "request"
	}
	fmt.Fprintf(output, "\tvar response %s\n", responseType)
	fmt.Fprintf(output, "\terr := client.execute(ctx, %sOperation, path, query, %s, &response, options...)\n", operation.ID, bodyExpression)
	if strings.HasPrefix(responseType, "[]") {
		fmt.Fprintln(output, "\treturn response, err")
	} else {
		fmt.Fprintln(output, "\treturn &response, err")
	}
	fmt.Fprintln(output, "}")
	return nil
}

func (generator *generator) emitQuery(output *bytes.Buffer, parameter parameter) {
	field := "params." + goName(parameter.Name)
	value := field
	isPointer := !parameter.Required || isNullable(parameter.Schema)
	if isPointer {
		fmt.Fprintf(output, "\tif %s != nil {\n", field)
		if primitiveType(parameter.Schema) != "date-time" {
			value = "*" + field
		}
	}
	switch primitiveType(parameter.Schema) {
	case "integer":
		fmt.Fprintf(output, "\tquery.Set(%q, strconv.FormatInt(int64(%s), 10))\n", parameter.Name, value)
	case "boolean":
		fmt.Fprintf(output, "\tquery.Set(%q, strconv.FormatBool(bool(%s)))\n", parameter.Name, value)
	case "date-time":
		fmt.Fprintf(output, "\tquery.Set(%q, %s.Format(time.RFC3339))\n", parameter.Name, value)
	default:
		fmt.Fprintf(output, "\tquery.Set(%q, string(%s))\n", parameter.Name, value)
	}
	if isPointer {
		fmt.Fprintln(output, "\t}")
	}
}

func (generator *generator) operations() ([]operation, error) {
	var operations []operation
	for path, pathValue := range generator.document.Paths {
		pathParameters := parameters(pathValue["parameters"])
		for method, raw := range pathValue {
			if !isHTTPMethod(method) {
				continue
			}
			value, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			id, _ := value["operationId"].(string)
			if id == "" {
				return nil, fmt.Errorf("%s %s has no operationId", strings.ToUpper(method), path)
			}
			if _, exists := generator.operationNames[id]; exists {
				return nil, fmt.Errorf("duplicate operationId %s", id)
			}
			generator.operationNames[id] = struct{}{}
			allParameters := append(pathParameters, parameters(value["parameters"])...)
			entry := operation{ID: id, Method: strings.ToUpper(method), Path: path, Idempotency: idempotency(value, strings.ToUpper(method), allParameters)}
			for _, parameter := range allParameters {
				switch parameter.Location {
				case "path":
					entry.PathParameters = append(entry.PathParameters, parameter)
				case "query":
					entry.QueryParameters = append(entry.QueryParameters, parameter)
				}
			}
			sort.Slice(entry.PathParameters, func(i, j int) bool { return entry.PathParameters[i].Name < entry.PathParameters[j].Name })
			sort.Slice(entry.QueryParameters, func(i, j int) bool { return entry.QueryParameters[i].Name < entry.QueryParameters[j].Name })
			entry.Request = contentSchema(value["requestBody"])
			entry.Response, entry.SuccessStatuses = successResponse(value)
			operations = append(operations, entry)
		}
	}
	sort.Slice(operations, func(i, j int) bool { return operations[i].ID < operations[j].ID })
	return operations, nil
}

func parameters(raw any) []parameter {
	values, _ := raw.([]any)
	result := make([]parameter, 0, len(values))
	for _, rawValue := range values {
		value, _ := rawValue.(map[string]any)
		name, _ := value["name"].(string)
		location, _ := value["in"].(string)
		required, _ := value["required"].(bool)
		schemaValue, _ := value["schema"].(map[string]any)
		result = append(result, parameter{Name: name, Location: location, Required: required, Schema: schemaValue})
	}
	return result
}

func contentSchema(raw any) schema {
	container, _ := raw.(map[string]any)
	content, _ := container["content"].(map[string]any)
	for _, contentType := range []string{"application/json", "application/x-www-form-urlencoded"} {
		media, _ := content[contentType].(map[string]any)
		if value, ok := media["schema"].(map[string]any); ok {
			return value
		}
	}
	return nil
}

func successResponse(operation map[string]any) (schema, []int) {
	responses, _ := operation["responses"].(map[string]any)
	var statuses []int
	for rawStatus := range responses {
		status, err := strconv.Atoi(rawStatus)
		if err == nil && status >= 200 && status < 300 {
			statuses = append(statuses, status)
		}
	}
	sort.Ints(statuses)
	if len(statuses) == 0 {
		return nil, nil
	}
	return contentSchema(responses[strconv.Itoa(statuses[0])]), statuses
}

func idempotency(operation map[string]any, method string, parameters []parameter) string {
	if value, ok := operation["x-aether-idempotency"].(string); ok {
		return value
	}
	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		return "not_applicable"
	}
	for _, parameter := range parameters {
		if strings.EqualFold(parameter.Name, "idempotency-key") {
			if parameter.Required {
				return "required"
			}
			return "optional"
		}
	}
	return "unsupported"
}

func (generator *generator) goType(value schema, contextName, currentName string) (string, bool, error) {
	if value == nil {
		return "any", false, fmt.Errorf("missing schema")
	}
	if name := generator.matchingName(value); name != "" && name != currentName {
		return name, isNullable(value), nil
	}
	if _, ok := value["const"]; ok {
		return "string", false, nil
	}
	if variants, ok := value["oneOf"].([]any); ok {
		return generator.nullableVariant(variants, contextName, currentName)
	}
	if variants, ok := value["anyOf"].([]any); ok {
		return generator.nullableVariant(variants, contextName, currentName)
	}
	if values := enumValues(value); values != nil {
		if name := generator.inlineEnums[enumKey(value)]; name != "" {
			return name, false, nil
		}
		return "string", false, nil
	}
	typeName, nullable := schemaTypeName(value)
	switch typeName {
	case "string":
		if value["format"] == "date-time" {
			return "time.Time", nullable, nil
		}
		return "string", nullable, nil
	case "integer":
		return "int64", nullable, nil
	case "number":
		return "float64", nullable, nil
	case "boolean":
		return "bool", nullable, nil
	case "array":
		items, _ := value["items"].(map[string]any)
		itemType, _, err := generator.goType(items, contextName+"Item", currentName)
		return "[]" + itemType, nullable, err
	case "object":
		if value["additionalProperties"] == true {
			return "map[string]any", nullable, nil
		}
		if itemSchema, ok := value["additionalProperties"].(map[string]any); ok {
			itemType, _, err := generator.goType(itemSchema, contextName+"Value", currentName)
			return "map[string]" + itemType, nullable, err
		}
		return generator.anonymousStruct(value, contextName, currentName)
	default:
		if value["allOf"] != nil {
			return generator.anonymousStruct(value, contextName, currentName)
		}
		return "any", nullable, fmt.Errorf("unsupported schema type %v", value["type"])
	}
}

func (generator *generator) nullableVariant(variants []any, contextName, currentName string) (string, bool, error) {
	var concrete []schema
	for _, raw := range variants {
		variant, _ := raw.(map[string]any)
		if typeName, _ := schemaTypeName(variant); typeName == "null" {
			continue
		}
		concrete = append(concrete, variant)
	}
	if len(concrete) != 1 || len(concrete) == len(variants) {
		return "", false, fmt.Errorf("unsupported polymorphism")
	}
	typeName, _, err := generator.goType(concrete[0], contextName, currentName)
	return typeName, true, err
}

func (generator *generator) anonymousStruct(value schema, contextName, currentName string) (string, bool, error) {
	properties, required, err := flattenedProperties(value)
	if err != nil {
		return "", false, err
	}
	var output strings.Builder
	output.WriteString("struct { ")
	for _, propertyName := range sortedKeys(properties) {
		property := properties[propertyName]
		typeName, nullable, err := generator.goType(property, contextName+goName(propertyName), currentName)
		if err != nil {
			return "", false, err
		}
		_, isRequired := required[propertyName]
		if nullable || !isRequired {
			typeName = pointerType(typeName)
		}
		tag := propertyName
		if !isRequired {
			tag += ",omitempty"
		}
		fmt.Fprintf(&output, "%s %s `json:%q`; ", goName(propertyName), typeName, tag)
	}
	output.WriteString("}")
	return output.String(), false, nil
}

func flattenedProperties(value schema) (map[string]schema, map[string]struct{}, error) {
	properties := make(map[string]schema)
	required := make(map[string]struct{})
	if variants, ok := value["allOf"].([]any); ok {
		for _, raw := range variants {
			variant, _ := raw.(map[string]any)
			variantProperties, variantRequired, err := flattenedProperties(variant)
			if err != nil {
				return nil, nil, err
			}
			for name, property := range variantProperties {
				if existing, exists := properties[name]; exists {
					if canonical(existing) != canonical(property) {
						return nil, nil, fmt.Errorf("conflicting allOf property %s", name)
					}
					continue
				}
				properties[name] = property
			}
			for name := range variantRequired {
				required[name] = struct{}{}
			}
		}
		return properties, required, nil
	}
	rawProperties, _ := value["properties"].(map[string]any)
	for name, raw := range rawProperties {
		property, _ := raw.(map[string]any)
		properties[name] = property
	}
	for _, raw := range stringSlice(value["required"]) {
		required[raw] = struct{}{}
	}
	return properties, required, nil
}

func (generator *generator) collectInlineEnums(value schema, contextName string) {
	if value == nil {
		return
	}
	if enumValues(value) != nil && generator.matchingName(value) == "" {
		key := enumKey(value)
		if _, exists := generator.inlineEnums[key]; !exists {
			generator.inlineEnums[key] = contextName
		}
	}
	properties, _, _ := flattenedProperties(value)
	for _, propertyName := range sortedKeys(properties) {
		candidate := goName(propertyName)
		if candidate == "Status" && strings.Join(enumValues(properties[propertyName]), ",") == "active,disabled" {
			candidate = "ResourceStatus"
		}
		generator.collectInlineEnums(properties[propertyName], candidate)
	}
	if items, ok := value["items"].(map[string]any); ok {
		generator.collectInlineEnums(items, contextName+"Item")
	}
	for _, keyword := range []string{"oneOf", "anyOf"} {
		variants, _ := value[keyword].([]any)
		for _, raw := range variants {
			variant, _ := raw.(map[string]any)
			generator.collectInlineEnums(variant, contextName)
		}
	}
}

func (generator *generator) matchingName(value schema) string {
	return generator.namedBySchema[canonical(value)]
}

func (generator *generator) sortedInlineEnumNames() []string {
	values := make([]string, 0, len(generator.inlineEnums))
	for _, name := range generator.inlineEnums {
		values = append(values, name)
	}
	sort.Strings(values)
	return values
}

func enumValues(value schema) []string {
	raw, ok := value["enum"].([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		text, ok := item.(string)
		if !ok {
			return nil
		}
		values = append(values, text)
	}
	return values
}

func schemaTypeName(value schema) (string, bool) {
	switch raw := value["type"].(type) {
	case string:
		return raw, false
	case []any:
		var concrete string
		nullable := false
		for _, item := range raw {
			typeName, _ := item.(string)
			if typeName == "null" {
				nullable = true
			} else if concrete == "" {
				concrete = typeName
			} else {
				return "", nullable
			}
		}
		return concrete, nullable
	default:
		if value["properties"] != nil {
			return "object", false
		}
		return "", false
	}
}

func primitiveType(value schema) string {
	if variants, ok := value["oneOf"].([]any); ok {
		for _, raw := range variants {
			variant, _ := raw.(map[string]any)
			if kind := primitiveType(variant); kind != "null" {
				return kind
			}
		}
	}
	typeName, _ := schemaTypeName(value)
	if typeName == "string" && value["format"] == "date-time" {
		return "date-time"
	}
	return typeName
}

func isNullable(value schema) bool {
	if _, nullable := schemaTypeName(value); nullable {
		return true
	}
	for _, keyword := range []string{"oneOf", "anyOf"} {
		variants, _ := value[keyword].([]any)
		for _, raw := range variants {
			variant, _ := raw.(map[string]any)
			if typeName, _ := schemaTypeName(variant); typeName == "null" {
				return true
			}
		}
	}
	return false
}

func isObject(value schema) bool {
	typeName, _ := schemaTypeName(value)
	return typeName == "object"
}

func canonical(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func enumKey(value schema) string {
	return canonical(map[string]any{"type": "string", "enum": enumValues(value)})
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringSlice(value any) []string {
	raw, _ := value.([]any)
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func goName(value string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	parts := strings.FieldsFunc(replacer.Replace(value), func(character rune) bool { return character == '_' || unicode.IsSpace(character) })
	for index, part := range parts {
		switch strings.ToLower(part) {
		case "id":
			parts[index] = "ID"
		case "url":
			parts[index] = "URL"
		case "uri":
			parts[index] = "URI"
		case "etag":
			parts[index] = "ETag"
		case "sha256":
			parts[index] = "SHA256"
		case "api":
			parts[index] = "API"
		case "http":
			parts[index] = "HTTP"
		case "cdn":
			parts[index] = "CDN"
		default:
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	name := strings.Join(parts, "")
	if strings.HasSuffix(name, "Id") {
		name = strings.TrimSuffix(name, "Id") + "ID"
	}
	return name
}

func exported(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	if value == "ID" {
		return "id"
	}
	if strings.HasSuffix(value, "ID") {
		return strings.ToLower(value[:len(value)-2]) + "ID"
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func parameterVariable(value string) string {
	name := lowerFirst(goName(value))
	if goKeywords[name] {
		return name + "Value"
	}
	return name
}

func pathValueExpression(parameter parameter) (string, error) {
	value := parameterVariable(parameter.Name)
	switch primitiveType(parameter.Schema) {
	case "string":
		return "string(" + value + ")", nil
	case "integer":
		return "strconv.FormatInt(int64(" + value + "), 10)", nil
	case "boolean":
		return "strconv.FormatBool(bool(" + value + "))", nil
	case "date-time":
		return value + ".Format(time.RFC3339)", nil
	default:
		return "", fmt.Errorf("unsupported path parameter %s type %q", parameter.Name, primitiveType(parameter.Schema))
	}
}

var goKeywords = map[string]bool{
	"break": true, "default": true, "func": true, "interface": true, "select": true,
	"case": true, "defer": true, "go": true, "map": true, "struct": true,
	"chan": true, "else": true, "goto": true, "package": true, "switch": true,
	"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
	"continue": true, "for": true, "import": true, "return": true, "var": true,
}

func pointerType(value string) string {
	if strings.HasPrefix(value, "*") {
		return value
	}
	return "*" + value
}

func methodConstant(method string) string {
	return "http.Method" + exported(strings.ToLower(method))
}

func idempotencyConstant(value string) string {
	switch value {
	case "required":
		return "transport.IdempotencyRequired"
	case "optional":
		return "transport.IdempotencyOptional"
	case "request_field":
		return "transport.IdempotencyRequestField"
	case "unsupported":
		return "transport.IdempotencyUnsupported"
	default:
		return "transport.IdempotencyNotApplicable"
	}
}

func isHTTPMethod(value string) bool {
	switch value {
	case "get", "post", "put", "patch", "delete", "options", "head", "trace":
		return true
	default:
		return false
	}
}

func fatalf(message string, arguments ...any) {
	fmt.Fprintf(os.Stderr, message+"\n", arguments...)
	os.Exit(1)
}
