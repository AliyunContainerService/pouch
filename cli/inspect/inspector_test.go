package inspect

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func TestNewTemplateInspector(t *testing.T) {
	type args struct {
		tmpl *template.Template
	}
	tests := []struct {
		name    string
		args    args
		want    Inspector
		wantOut string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			got, err := NewTemplateInspectorFromString(out, tt.args.tmplStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplateInspectorFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
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
		references string
		tmplStr    string
		getRef     GetRefFunc
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "testInspectDefault",
			args: args{
				references: "test",
				tmplStr:    "",
				getRef:     getRefFunc,
			},
			wantOut: "id",
			wantErr: false,
		}, {
			name: "testInspectTemplate",
			args: args{
				references: "test",
				tmplStr:    "{{.ID}}",
				getRef:     getRefFunc,
			},
			wantOut: "id",
			wantErr: false,
		}, {
			name: "testInspectTemplateError",
			args: args{
				references: "test",
				tmplStr:    "{{.Id}}",
				getRef:     getRefFunc,
			},
			wantOut: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			if err := Inspect(out, []string{tt.args.references}, tt.args.tmplStr, tt.args.getRef); (err != nil) != tt.wantErr {
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
