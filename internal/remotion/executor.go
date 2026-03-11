package remotion

import "context"

// Executor defines how to interact with the external remotion CLI.
type Executor interface {
	// Render triggers `remotion render` with a given props JSON file,
	// rendering out to outputPath.
	Render(ctx context.Context, templatePath string, composition string, propsPath string, outputPath string) error

	// Preview triggers `remotion studio` to allow live-previewing.
	Preview(ctx context.Context, templatePath string, composition string, propsPath string) error
}
