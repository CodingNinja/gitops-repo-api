package entrypoint

// func ExtractResources(root string, entrypoint Entrypoint) (interface{}, error) {
// 	if entrypoint.Type == EntrypointTypeKustomize {
// 		return resource.RenderKustomize(filepath.Join(root, entrypoint.Directory))
// 	}

// 	if entrypoint.Type == EntrypointTypeTerraform {
// 		return resource.RenderTerraform(filepath.Join(root, entrypoint.Directory))
// 	}

// 	return nil, fmt.Errorf("unable to extract resources from unknown entrypoint type %q", entrypoint.Type)
// }
