# model-catalog

This Golang type definition is automatically generated from the [model catalog schema](../model-catalog-schema.json) using [quicktype](https://github.com/glideapps/quicktype).

## Updating 

If you are making changes to the schema and need to re-generate the types, run `make generate-schema-golang` from the root of the repository. 

## Type Generation

The command that `make generate-schema-golang` runs to generate the types is 
```
cd schema; sed 's|\#/$$defs/modelServerAPI|\#/$$defs/modelServer/$$defs/modelServerAPI|g' model-catalog.schema.json | quicktype -s schema -o types/golang/model-catalog.go --package golang
```

**Note:** Due to implementation differences between different JSON Schema parsers, some parsers treat references against nested names as relative references, and others absolute. Quicktype requires absolute references, which is why we must use `sed` to change our reference to `modelServerAPI` from relative to absolute.