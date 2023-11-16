# protobuf-go-plugins

Useful protoc-gen plugins for go and a starting point for writing new custom plugins.

## Enum-desc-go
Annotate enum values with human-friendly display strings and generate to maps.

```proto
enum Enum {
    OPTION_A = 0 [
        (kazweb.annotations.desc) = "Plan A"
    ];
    OPTION_B = 1 [
        (kazweb.annotations.desc) = "Plan B"
    ];
}
```

will generate:
```go
var Enum_desc = map[int32]string{
	0: "Plan A",
	1: "Plan B",
}
```

## Sam-go
Generate a Single-Abstract-Method interface for any annotated protobuf service method.  

```proto
service SampleService {
    rpc Echo(EchoRequest) returns (EchoResponse) {
        option (kazweb.annotations.abstract) = true;
    };
}
```

produces:
```go
type CanEcho interface {
	Echo(ctx context.Context, in *EchoRequest) (*EchoResponse, error)
}

type EchoFunc func(ctx context.Context, in *EchoRequest) (*EchoResponse, error)
```