package integritycheck

import (
	"encoding/json"
	"fmt"
	"golang.org/x/tools/go/analysis/analysistest"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"reflect"
	"testing"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
func TestDecodeFileDesc(t *testing.T) {
	var testFD  = []byte{
		0x0a, 0x29, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x65,
		0x65, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x5f, 0x65, 0x78,
		0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x65, 0x78, 0x61,
		0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x69,
		0x74, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x69, 0x6e,
		0x74, 0x65, 0x67, 0x72, 0x69, 0x74, 0x79, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
		0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x38, 0x0a, 0x0f, 0x53, 0x74, 0x65, 0x65, 0x72,
		0x69, 0x6e, 0x67, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x74,
		0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x61, 0x6e, 0x67, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01,
		0x28, 0x02, 0x52, 0x0d, 0x73, 0x74, 0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x41, 0x6e, 0x67, 0x6c,
		0x65, 0x22, 0x62, 0x0a, 0x1b, 0x53, 0x74, 0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x43, 0x6f, 0x6d,
		0x6d, 0x61, 0x6e, 0x64, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
		0x12, 0x25, 0x0a, 0x0e, 0x73, 0x74, 0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x61, 0x6e, 0x67,
		0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0d, 0x73, 0x74, 0x65, 0x65, 0x72, 0x69,
		0x6e, 0x67, 0x41, 0x6e, 0x67, 0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61,
		0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e,
		0x61, 0x74, 0x75, 0x72, 0x65, 0x22, 0x6f, 0x0a, 0x21, 0x53, 0x74, 0x65, 0x65, 0x72, 0x69, 0x6e,
		0x67, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61,
		0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x74,
		0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x61, 0x6e, 0x67, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01,
		0x28, 0x02, 0x52, 0x0d, 0x73, 0x74, 0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x41, 0x6e, 0x67, 0x6c,
		0x65, 0x12, 0x23, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02,
		0x20, 0x01, 0x28, 0x0c, 0x42, 0x05, 0x92, 0x44, 0x02, 0x10, 0x02, 0x52, 0x09, 0x73, 0x69, 0x67,
		0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x42, 0x45, 0x5a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
		0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x69, 0x6e, 0x72, 0x69, 0x64, 0x65, 0x2f, 0x70, 0x72, 0x6f,
		0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x69,
		0x6e, 0x74, 0x65, 0x67, 0x72, 0x69, 0x74, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67,
		0x65, 0x6e, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70,
		0x72, 0x6f, 0x74, 0x6f, 0x33,
	}
	expectedProtoMap := map[string]struct{} {
		"SteeringCommandVerificationOption": *new(struct{}),
	}
	expectedFileName := protoreflect.Name("v1")

	fd := protoimpl.DescBuilder{RawDescriptor: testFD}.Build().File
	if fd.Name() != expectedFileName {
		t.Errorf("Error actual name: %v expected: %v", fd.Name(), expectedFileName)
	}
	protoMap := findSigRequiredProtos(fd)
	for _, n := range protoMap {
		fmt.Println(n)
	}
	if  !reflect.DeepEqual(protoMap, expectedProtoMap) {
		b, err := json.MarshalIndent(protoMap, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		be, err := json.MarshalIndent(expectedProtoMap, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		t.Errorf("signature not enabled for protos were %v, expected: %v", string(b), string(be))
	}
}

