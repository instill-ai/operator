//go:build integration
// +build integration

package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/operator/pkg/base64"
	"github.com/instill-ai/operator/pkg/rest"
	"github.com/instill-ai/operator/pkg/textextraction"

	jsonop "github.com/instill-ai/operator/pkg/json"
)

var (
	base64Oper       base.IExecution
	textExtractionOp base.IExecution
	jsonOp           base.IExecution
	restOp           base.IExecution
)

func init() {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{}}
	o := Init(nil, OperatorOptions{})
	base64Oper, _ = o.CreateExecution(o.ListOperatorDefinitionUids()[0], config, nil)
	textExtractionOp, _ = o.CreateExecution(o.ListOperatorDefinitionUids()[1], config, nil)
	jsonOp, _ = o.CreateExecution(o.ListOperatorDefinitionUids()[4], config, nil)
	restOp, _ = o.CreateExecution(o.ListOperatorDefinitionUids()[5], config, nil)
}

func TestBase64(t *testing.T) {
	file, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := base64.Base64{Data: string(file)}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_ENCODE"}}
	op, err := base64Oper.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)

	b, _ = json.Marshal(op)
	json.Unmarshal(b, &req)
	b, _ = json.Marshal(req)
	fmt.Printf("\n\n bytes: %s", req.Data)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_DECODE"}}
	op, err = base64Oper.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)

	b, _ = json.Marshal(op)
	json.Unmarshal(b, &req)
	ioutil.WriteFile("test_artifacts/image_res.jpg", []byte(req.Data), 0644)
}

func TestTextExtraction(t *testing.T) {
	path := "test_artifacts/resume.pdf"
	file, _ := ioutil.ReadFile(path)
	fileReq := textextraction.FromFile{
		FileContents: string(file),
		ContentType:  "application/pdf",
	}
	var in structpb.Struct
	b, _ := json.Marshal(fileReq)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_EXTRACT_FROM_FILE"}}
	op, err := textExtractionOp.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)

	pathReq := textextraction.FromPath{FilePath: path}
	b, _ = json.Marshal(pathReq)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_EXTRACT_FROM_PATH"}}
	op, err = textExtractionOp.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)

	webPath := "https://instill.tech"
	pathReq = textextraction.FromPath{FilePath: webPath}
	b, _ = json.Marshal(pathReq)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_EXTRACT_FROM_PATH"}}
	op, err = textExtractionOp.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}

func TestJSON(t *testing.T) {
	tests := []struct {
		input jsonop.GetValueInput
	}{
		{
			input: jsonop.GetValueInput{
				Path:       ".",
				JSONString: `{"a":{"b":{"c": [1, 2, 3, 4]}}}`,
			},
		},
		{
			input: jsonop.GetValueInput{
				Path:       "a",
				JSONString: `{"a":{"b":{"c": [1, 2, 3, 4]}}}`,
			},
		},
		{
			input: jsonop.GetValueInput{
				Path:       "a.b",
				JSONString: `{"a":{"b":{"c": [1, 2, 3, 4]}}}`,
			},
		},
		{
			input: jsonop.GetValueInput{
				Path:       "a.b.c[0]",
				JSONString: `{"a":{"b":{"c": [1, 2, 3, 4]}}}`,
			},
		},
	}
	var in structpb.Struct
	for _, test := range tests {
		t.Run(test.input.Path, func(t *testing.T) {
			b, _ := json.Marshal(test.input)
			protojson.Unmarshal(b, &in)
			in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_GET_VALUE"}}
			op, err := jsonOp.Execute([]*structpb.Struct{&in})
			fmt.Printf("\n op :%v, err:%s \n", op, err)
		})
	}
}

func TestREST(t *testing.T) {
	req := rest.Request{
		URL:         "https://httpbin.org/post",
		Method:      "POST",
		RequestBody: `{"hi": "hello"}`,
		Headers:     nil,
	}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_CALL_ENDPOINT"}}
	op, err := restOp.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)
}
