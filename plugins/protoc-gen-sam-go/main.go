package main

import (
	"flag"

	"github.com/aaronkaz/protobuf-go-plugins/internal/protoplugin"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/pluginpb"
)

const pluginName = "protoc-gen-sam-go"
const version = "0.1.0"
const extNumber = 200000001

func main() {
	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// The type information for all extensions is in the source files,
		// so we need to extract them into a dynamically created protoregistry.Types.
		extTypes := new(protoregistry.Types)
		for _, file := range gen.Files {
			if err := protoplugin.RegisterAllExtensions(extTypes, file.Desc); err != nil {
				panic(err)
			}
		}

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			mtds, err := getAbstractMethods(extTypes, f)
			if err != nil {
				panic(err)
			}

			generateFile(gen, f, mtds)
		}

		return nil
	})
}
