// Package integrity check defines an Analyzer that check that ProtoMessages Signatures are used
package integritycheck

import (
	"errors"
	"fmt"
	integritypb "github.com/einride/protoc-gen-messageintegrity/proto/gen/integrity/v1"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/loader"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
	"strconv"
	"strings"
)

var Analyzer = &analysis.Analyzer{
	Name: "integritylint",
	Doc:  "checks signatures are used if message integrity is enabled for protomessages",
	Run:  run,
}

// run steps involved.
// 1. Go through the imports and see if any of them are proto imports by checking recursively if they import the
// internal proto implementation packages.
// 2. Go through each struct proto found from the previous and see if any of them have the message integrity signature
// option set for a field.
// 		a. Check the extensions on the fields and see if any of them are message_integrity_signature
// 		b. See that it is REQUIRED
// 3. Make a short list of these struct types.
// 4. Go through the file looking for proto.Marshal() calls on any of these types.
// 5. I found look that proto.Sign() is called immediately before it.
// 4. Go through the file looking for proto.Unmarshal() calls on any of these types.
// 5. I found look that proto.Verify() is called immediately after it.

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Map of proto message types that have implicit message integrity enabled so they can be looked up when
		// Marshal and unmarshal calls are found.
		p := protoIntegrityChecker{typeInfo: pass.TypesInfo}
		protoTypes := make(map[types.Type]bool)
		ast.Inspect(file, func(n ast.Node) bool {
			im, ok := n.(*ast.ImportSpec)
			if !ok {
				return true
			}
			path, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				log.Fatal(err)
			}
			if strings.Compare(path, "command-line-arguments") == 0 {
				// I don't know why this always appears as an import but I'm ignoring it.
				return true
			}
			var conf loader.Config
			conf.Import(path)
			prog, err := conf.Load()
			if err != nil {
				log.Fatal(err)
			}
			for pn := range prog.AllPackages {
				// Weak way of figuring out if an import is a proto message generated from protoc-go
				// I only look for proto message structs that have the options in dependencies which have protoimpl as
				// a dependency to save time.
				if pn.Path() != "google.golang.org/protobuf/runtime/protoimpl" {
					continue
				}
				integrityProtosList, err := p.findIntegrityProtos(prog.InitialPackages())
				if err != nil {
					log.Fatal(err)
				}
				for _, t := range integrityProtosList {
					protoTypes[t] = true
					to := p.typeInfo.ObjectOf(im.Name)
					pass.ExportObjectFact(to, &integrityFact{t: t})
				}
			}
			return true
		})
		// Now I have a list of all of the types find Un-marshals of.
		ast.Inspect(file, parseUnmarshals)
	}
	return nil, nil
}

type protoIntegrityChecker struct {
	typeInfo *types.Info
}

func (p *protoIntegrityChecker) findIntegrityProtos(packages []*loader.PackageInfo) ([]types.Type, error) {
	var integrityProtoTypes []types.Type
	for _, p := range packages {
		for _, f := range p.Files {
			fmt.Println(f.Scope)
			for _, d := range f.Scope.Objects {
				if d.Kind != ast.Var {
					continue
				}
				// This really should be "file_example_" + <import_file_name> + "_.*_proto_rawDesc"
				// but I don't know how to get the file name of the proto file name
				if !strings.Contains(d.Name, "_proto_rawDesc") {
					continue
				}
				if d.Decl == nil {
					continue
				}
				as, ok := d.Decl.(*ast.ValueSpec)
				if !ok {
					fmt.Println("not assignment statement")
					continue
				}
				for _, v := range as.Values {
					fieldDescRaw, err := parseByteSlice(v)
					if err != nil {
						continue
					}
					fd := protoimpl.DescBuilder{RawDescriptor: fieldDescRaw}.Build().File
					if fd == nil {
						log.Fatal("failed to parse raw field desc")
					}
					protosMap := findSigRequiredProtos(fd)
					for k := range protosMap {
						// Map from a string of the type name from the proto file descriptor to the
						// types.Type representation for analysis.
						to, exists := f.Scope.Objects[k]
						if !exists {
							return nil, fmt.Errorf("proto Message type %v not found in scope", k)
						}
						ts, ok := to.Decl.(*ast.TypeSpec)
						if !ok {
							return nil, errors.New("decl could not be converted to a typeSpec")
						}
						t := p.TypeOf(ts.Type)
						if !ok {
							return nil, fmt.Errorf("proto Message object %v could not be converted to a type", k)
						}
						integrityProtoTypes = append(integrityProtoTypes, t)
					}
				}
			}
		}
	}
	return integrityProtoTypes, nil
}

// Find Unmarshals of integrity enabled protos.
func parseUnmarshals(n ast.Node) bool {
	return true
}

// find the protos that have the message integrity signature option set for a field.
func findSigRequiredProtos(fd protoreflect.FileDescriptor) map[string]struct{} {
	protoMap := make(map[string]struct{})
	for i := 0; i < fd.Messages().Len(); i++ {
		m := fd.Messages().Get(i)
		fields := m.Fields()
		hasOption := false
		for i := 0; i < fields.Len(); i++ {
			fd := fields.Get(i)
			opts := fd.Options().(*descriptorpb.FieldOptions)
			sigOption, ok := proto.GetExtension(opts, integritypb.E_Signature).(*integritypb.Signature)
			if !ok || fd.Kind() != protoreflect.BytesKind {
				continue // The signature can only be a bytes field, ignore everything else.
			} else if sigOption.GetBehaviour() != integritypb.SignatureBehaviour_SIGNATURE_BEHAVIOUR_REQUIRED &&
				sigOption.GetBehaviour() != integritypb.SignatureBehaviour_SIGNATURE_BEHAVIOUR_OPTIONAL {
				// Don't verify if the option isn't there or it's behaviour isn't required or optional.
				continue
			}
			hasOption = true
		}
		if hasOption {
			s := string(m.Name())
			protoMap[s] = *new(struct{})
		}
	}
	return protoMap
}

// reads a constant byte array in a ast.Composite lit and turns it into a runtime byte array.
func parseByteSlice(v ast.Expr) ([]byte, error) {
	// Sneakily I can parse this byte array into the field descriptor proto at run time.
	cl, ok := v.(*ast.CompositeLit)
	if !ok {
		return nil, errors.New("not a comp lit")
	}
	if cl.Elts == nil {
		return nil, errors.New("no elements")
	}
	var fieldDescRaw []byte
	for _, e := range cl.Elts {
		v, ok := e.(*ast.BasicLit)
		if !ok {
			continue
		}
		// I feel like this should work with bitset=8 but I get failed to parse 0x92 so 16
		// bits it is.
		ib, err := strconv.ParseInt(v.Value, 0, 16)
		if err != nil {
			return nil, err
		}
		b := byte(ib)
		fieldDescRaw = append(fieldDescRaw, b)
	}
	return fieldDescRaw, nil
}

// integrityFact is a fact associated with types that are Protocol buffer Messages with the message integrity
// option enabled.
type integrityFact struct {
	t types.Type
}

func (i *integrityFact) String() string {
	return fmt.Sprintf("found proto message type: %v... with message integrity enabled", i.t.String()[:70])
}
func (*integrityFact) AFact() {}
