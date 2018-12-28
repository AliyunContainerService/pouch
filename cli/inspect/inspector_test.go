package inspect

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/alibaba/pouch/pkg/utils/templates"

	"github.com/google/go-cmp/cmp"
)

func TestNewTemplateInspector(t *testing.T) {
	// Prepare test data
	idTmpl, _ := template.New("idTemplate").Parse("{{.ID}}")
	type args struct {
		tmpl *template.Template
	}
	tests := []struct {
		name    string
		args    args
		want    Inspector
		wantOut string
	}{
		{
			name: "testTemplateInspector",
			args: args{
				tmpl: idTmpl,
			},
			want: &TemplateInspector{
				outputStream: bytes.NewBufferString("pouch"),
				buffer:       new(bytes.Buffer),
				tmpl:         idTmpl,
			},
			wantOut: "pouch",
		},
	}
	for _, tt := range tests {
		out := bytes.NewBuffer([]byte("pouch"))
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTemplateInspector(out, tt.args.tmpl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTemplateInspector() = %v, want %v", got, tt.want)
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("NewTemplateInspector() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func TestNewTemplateInspectorFromString(t *testing.T) {
	// Prepare test data
	idTmpl, _ := templates.Parse("{{.ID}}")
	type args struct {
		tmplStr string
	}
	tests := []struct {
		name    string
		args    args
		want    Inspector
		wantOut string
		wantErr bool
	}{
		{
			name: "testEmptyTmplStr",
			args: args{
				tmplStr: "",
			},
			want: &IndentedInspector{
				outputStream: &bytes.Buffer{},
				elements:     nil,
				rawElements:  nil,
			},
			wantOut: "",
			wantErr: false,
		},
		{
			name: "testCorrectTmplStr",
			args: args{
				tmplStr: "{{.ID}}",
			},
			want: &TemplateInspector{
				outputStream: &bytes.Buffer{},
				buffer:       new(bytes.Buffer),
				tmpl:         idTmpl,
			},
			wantOut: "",
			wantErr: false,
		},
		{
			name: "testErrorTmplStr",
			args: args{
				tmplStr: "{{xxx}}",
			},
			want:    nil,
			wantOut: "",
			wantErr: true,
		},
	}
	opts := cmp.Options{
		// This option declares approximate equality on IndentedInspector
		cmp.Comparer(func(x, y IndentedInspector) bool {
			return reflect.DeepEqual(x, y)
		}),
		// This option declares approximate equality on TemplateInspector
		cmp.Comparer(func(x, y TemplateInspector) bool {
			return (reflect.DeepEqual(x.outputStream, y.outputStream)) &&
				(reflect.DeepEqual(x.buffer, y.buffer)) &&
				(reflect.DeepEqual(x.tmpl.Name(), y.tmpl.Name())) &&
				(reflect.DeepEqual(x.tmpl.Tree.Root.Nodes, y.tmpl.Tree.Root.Nodes))
		}),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			got, err := NewTemplateInspectorFromString(out, tt.args.tmplStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplateInspectorFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// There are some unexported fields inside template ,
			// reflect.DeepEqual is not suitable for template comparation
			// so use cmp.Equal instead
			if !cmp.Equal(got, tt.want, opts) {
				t.Errorf("NewTemplateInspectorFromString() = %v, want %v", got, tt.want)
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("NewTemplateInspectorFromString() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func TestInspect(t *testing.T) {
	// Prepare test data
	type testElement struct {
		ID string `json:"Id"`
	}

	getRefFunc := func(ref string) (interface{}, error) {
		return testElement{
			ID: "id",
		}, nil
	}

	type args struct {
		references []string
		tmplStr    string
		getRef     GetRefFunc
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		{
			name: "testInspectEmptyTmplStr",
			args: args{
				references: []string{"single reference"},
				tmplStr:    "",
				getRef:     getRefFunc,
			},
			wantOut: "id",
			wantErr: false,
		}, {
			name: "testInspectTemplateSingleReference",
			args: args{
				references: []string{"single reference"},
				tmplStr:    "{{.ID}}",
				getRef:     getRefFunc,
			},
			wantOut: "id",
			wantErr: false,
		}, {
			name: "testInspectTemplateMultiReferences",
			args: args{
				references: []string{"reference1", "reference2"},
				tmplStr:    "{{.ID}}",
				getRef:     getRefFunc,
			},
			wantOut: "id",
			wantErr: false,
		}, {
			name: "testInspectTemplateEmptyReference",
			args: args{
				references: []string{},
				tmplStr:    "{{.ID}}",
				getRef:     getRefFunc,
			},
			wantOut: "\n",
			wantErr: false,
		}, {
			name: "testInspectTemplateError",
			args: args{
				references: []string{"single reference"},
				tmplStr:    "{{.Id}}",
				getRef:     getRefFunc,
			},
			wantOut: "",
			wantErr: true,
		}, {
			name: "testInspectTemplateError2",
			args: args{
				references: []string{"reference1", "reference2"},
				tmplStr:    "{{.NotExists}}",
				getRef:     getRefFunc,
			},
			wantOut: "",
			wantErr: true,
		}, {
			name: "testInspectTemplateError3",
			args: args{
				references: []string{"single reference"},
				tmplStr:    "{{.Unclosed}",
				getRef:     getRefFunc,
			},
			wantOut: "",
			wantErr: true,
		}, {
			name: "testInspectTemplateError4",
			args: args{
				references: []string{},
				tmplStr:    "{{NotExistsFunction .WhenEmptyReference}}",
				getRef:     getRefFunc,
			},
			wantOut: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			if err := Inspect(out, tt.args.references, tt.args.tmplStr, tt.args.getRef); (err != nil) != tt.wantErr {
				t.Errorf("Inspect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotOut := out.String(); !strings.Contains(gotOut, tt.wantOut) {
				t.Errorf("Inspect() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func TestTemplateInspector_Inspect(t *testing.T) {
	// Prepare test data
	idTmpl, _ := templates.Parse("{{.ID}}")
	tmplErr, _ := templates.Parse("{{.Id}}")
	type testElement struct {
		ID string `json:"Id"`
	}
	type fields struct {
		outputStream io.Writer
		buffer       *bytes.Buffer
		tmpl         *template.Template
	}
	type args struct {
		typedElement interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "testTemplateInspectorInspect",
			fields: fields{
				outputStream: &bytes.Buffer{},
				buffer:       new(bytes.Buffer),
				tmpl:         idTmpl,
			},
			args: args{
				typedElement: &testElement{
					ID: "id",
				},
			},
			wantErr: false,
		}, {
			name: "testTemplateInspectorInspectError",
			fields: fields{
				outputStream: &bytes.Buffer{},
				buffer:       new(bytes.Buffer),
				tmpl:         tmplErr,
			},
			args: args{
				typedElement: &testElement{
					ID: "id",
				},
			},
			wantErr: true,
		}, {
			name: "testTemplateInspectorInspectNilOutputStream",
			fields: fields{
				outputStream: nil,
				buffer:       new(bytes.Buffer),
				tmpl:         idTmpl,
			},
			args: args{
				typedElement: &testElement{
					ID: "id",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &TemplateInspector{
				outputStream: tt.fields.outputStream,
				buffer:       tt.fields.buffer,
				tmpl:         tt.fields.tmpl,
			}
			if err := i.Inspect(tt.args.typedElement); (err != nil) != tt.wantErr {
				t.Errorf("TemplateInspector.Inspect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateInspector_Flush(t *testing.T) {
	// Prepare test data
	idTmpl, _ := templates.Parse("{{.ID}}")
	buf := bytes.NewBuffer([]byte("some content"))
	type fields struct {
		outputStream io.Writer
		buffer       *bytes.Buffer
		tmpl         *template.Template
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "testTemplateInspectorFlushBufferLenZero",
			fields: fields{
				outputStream: &bytes.Buffer{},
				buffer:       new(bytes.Buffer),
				tmpl:         idTmpl,
			},
			wantErr: false,
		}, {
			name: "testTemplateInspectorFlushBufferLenNotZero",
			fields: fields{
				outputStream: buf,
				buffer:       new(bytes.Buffer),
				tmpl:         idTmpl,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &TemplateInspector{
				outputStream: tt.fields.outputStream,
				buffer:       tt.fields.buffer,
				tmpl:         tt.fields.tmpl,
			}
			if err := i.Flush(); (err != nil) != tt.wantErr {
				t.Errorf("TemplateInspector.Flush() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewIndentedInspector(t *testing.T) {
	tests := []struct {
		name             string
		want             Inspector
		wantOutputStream string
	}{
		{
			name: "testIndentedInspector",
			want: &IndentedInspector{
				outputStream: &bytes.Buffer{},
				elements:     nil,
				rawElements:  nil,
			},
			wantOutputStream: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputStream := &bytes.Buffer{}
			if got := NewIndentedInspector(outputStream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIndentedInspector() = %v, want %v", got, tt.want)
			}
			if gotOutputStream := outputStream.String(); gotOutputStream != tt.wantOutputStream {
				t.Errorf("NewIndentedInspector() = %v, want %v", gotOutputStream, tt.wantOutputStream)
			}
		})
	}
}

func TestIndentedInspector_Inspect(t *testing.T) {
	// Prepare test data
	var dataSlice [2]string
	dataSlice[0] = "Hello"
	dataSlice[1] = "Pouch"
	interfaceSlice := make([]interface{}, len(dataSlice))
	for i, d := range dataSlice {
		interfaceSlice[i] = d
	}
	rawElements := [][]byte{
		[]byte("Hello"),
		[]byte("Pouch"),
	}
	type testElement struct {
		ID string `json:"Id"`
	}
	type fields struct {
		outputStream io.Writer
		elements     []interface{}
		rawElements  [][]byte
	}
	type args struct {
		typedElement interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "testInspect",
			fields: fields{
				outputStream: &bytes.Buffer{},
				elements:     interfaceSlice,
				rawElements:  rawElements,
			},
			args: args{
				typedElement: &testElement{
					ID: "id",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IndentedInspector{
				outputStream: tt.fields.outputStream,
				elements:     tt.fields.elements,
				rawElements:  tt.fields.rawElements,
			}
			if err := i.Inspect(tt.args.typedElement); (err != nil) != tt.wantErr {
				t.Errorf("IndentedInspector.Inspect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIndentedInspector_Flush(t *testing.T) {
	// Prepare test data
	var dataSlice [2]string
	dataSlice[0] = "Hello"
	dataSlice[1] = "Pouch"
	interfaceSlice := make([]interface{}, len(dataSlice))
	for i, d := range dataSlice {
		interfaceSlice[i] = d
	}
	rawElements := [][]byte{
		[]byte("Hello"),
		[]byte("Pouch"),
	}
	type fields struct {
		outputStream io.Writer
		elements     []interface{}
		rawElements  [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "testIndentedInspectorFlush",
			fields: fields{
				outputStream: &bytes.Buffer{},
				elements:     interfaceSlice,
				rawElements:  rawElements,
			},
			wantErr: false,
		}, {
			name: "testIndentedInspectorFlushElementsNil",
			fields: fields{
				outputStream: &bytes.Buffer{},
				elements:     nil,
				rawElements:  rawElements,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &IndentedInspector{
				outputStream: tt.fields.outputStream,
				elements:     tt.fields.elements,
				rawElements:  tt.fields.rawElements,
			}
			if err := i.Flush(); (err != nil) != tt.wantErr {
				t.Errorf("IndentedInspector.Flush() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
