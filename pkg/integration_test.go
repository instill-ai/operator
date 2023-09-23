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
)

var (
	oper base.IOperation
)

func init() {
	config := &structpb.Struct{
		Fields: map[string]*structpb.Value{}}
	o := Init(nil, OperatorOptions{})
	oper, _ = o.CreateOperation(o.ListOperatorDefinitionUids()[0], config, nil)
}

func TestBase64(t *testing.T) {
	file, _ := ioutil.ReadFile("test_artifacts/image.jpg")
	req := base64.Base64{Data: string(file)}
	var in structpb.Struct
	b, _ := json.Marshal(req)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_ENCODE"}}
	op, err := oper.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)

	b, _ = json.Marshal(op)
	json.Unmarshal(b, &req)
	b, _ = json.Marshal(req)
	fmt.Printf("\n\n bytes: %s", req.Data)
	protojson.Unmarshal(b, &in)
	in.Fields["task"] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "TASK_DECODE"}}
	op, err = oper.Execute([]*structpb.Struct{&in})
	fmt.Printf("\n op :%v, err:%s", op, err)

	b, _ = json.Marshal(op)
	json.Unmarshal(b, &req)
	ioutil.WriteFile("test_artifacts/image_res.jpg", []byte(req.Data), 0644)
}
