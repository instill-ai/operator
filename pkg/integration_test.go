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
	"github.com/instill-ai/operator/pkg/textextraction"
)

var (
	base64Oper       base.IExecution
	textExtractionOp base.IExecution
)

func init() {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{}}
	o := Init(nil, OperatorOptions{})
	base64Oper, _ = o.CreateExecution(o.ListOperatorDefinitionUids()[0], config, nil)
	textExtractionOp, _ = o.CreateExecution(o.ListOperatorDefinitionUids()[1], config, nil)
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
