package llm

import "github.com/invopop/jsonschema"

func GenerateSchema(v any) any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	return reflector.Reflect(v)
}