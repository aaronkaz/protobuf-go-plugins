package protoplugin

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type FileGenerator func(*protogen.Plugin, *protogen.File, *protoregistry.Types) *protogen.GeneratedFile

func Run(generator FileGenerator) {
	var flags flag.FlagSet

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// The type information for all extensions is in the source files,
		// so we need to extract them into a dynamically created protoregistry.Types.
		extTypes := new(protoregistry.Types)
		for _, file := range gen.Files {
			if err := RegisterAllExtensions(extTypes, file.Desc); err != nil {
				panic(err)
			}
		}

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			generator(gen, f, extTypes)
		}

		return nil
	})
}

func ProtocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

func OptionsReflect(extTypes *protoregistry.Types, options interface {
	Reset()
	ProtoReflect() protoreflect.Message
}) (protoreflect.Message, error) {
	// The (Message|Service|Method|<*>options as provided by protoc does not know about
	// dynamically created extensions, so they are left as unknown fields.
	// We round-trip marshal and unmarshal the options with
	// a dynamically created resolver that does know about extensions at runtime.
	b, err := proto.Marshal(options)
	if err != nil {
		return nil, err
	}
	options.Reset()
	err = proto.UnmarshalOptions{Resolver: extTypes}.Unmarshal(b, options)
	if err != nil {
		return nil, err
	}

	return options.ProtoReflect(), nil
}

func IsExtension(msg protoreflect.Message, ext int32) (match bool) {
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if !fd.IsExtension() {
			return true
		}

		if fd.Number() == protowire.Number(ext) {
			match = true
		}

		return true
	})

	return match
}

func GetProtoReflectFieldValue(msg protoreflect.Message, fieldName string) (val protoreflect.Value, err error) {
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		field := fd.Message().Fields().ByName(protoreflect.Name(fieldName))
		if field == nil {
			err = fmt.Errorf("'%s' field not found", fieldName)
			return true
		}
		val = v.Message().Get(field)
		return true
	})

	return
}

// Register all extensions into the provided protoregistry.Types
func RegisterAllExtensions(extTypes *protoregistry.Types, descs interface {
	Extensions() protoreflect.ExtensionDescriptors
}) error {
	xds := descs.Extensions()
	for i := 0; i < xds.Len(); i++ {
		if err := extTypes.RegisterExtension(dynamicpb.NewExtensionType(xds.Get(i))); err != nil {
			return err
		}
	}
	return nil
}
