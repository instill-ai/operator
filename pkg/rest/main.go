package rest

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/configLoader"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	callEndpoint = "TASK_CALL_ENDPOINT"
)

var (
	//go:embed config/definitions.json
	definitionJSON []byte
	once           sync.Once
	operator       base.IOperator
)

type OperatorOptions struct{}

type Operator struct {
	base.BaseOperator
	options OperatorOptions
}

type Operation struct {
	base.BaseExecution
	operator *Operator
}

func Init(logger *zap.Logger, options OperatorOptions) base.IOperator {
	once.Do(func() {
		loader := configLoader.InitJSONSchema(logger)
		connDefs, err := loader.LoadOperator(definitionJSON)
		if err != nil {
			panic(err)
		}
		operator = &Operator{
			BaseOperator: base.BaseOperator{Logger: logger},
			options:      options,
		}
		for idx := range connDefs {
			err := operator.AddOperatorDefinition(uuid.FromStringOrNil(connDefs[idx].GetUid()), connDefs[idx].GetId(), connDefs[idx])
			if err != nil {
				logger.Warn(err.Error())
			}
		}
	})
	return operator
}

func (o *Operator) CreateExecution(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	def, err := o.GetOperatorDefinitionByUid(defUid)
	if err != nil {
		return nil, err
	}
	return &Operation{
		BaseExecution: base.BaseExecution{
			Logger: logger, DefUid: defUid,
			Config:                config,
			OpenAPISpecifications: def.Spec.OpenapiSpecifications,
		},
		operator: o,
	}, nil
}

func (c *Operation) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := []*structpb.Struct{}
	task := inputs[0].GetFields()["task"].GetStringValue()
	for _, input := range inputs {
		if input.GetFields()["task"].GetStringValue() != task {
			return nil, fmt.Errorf("each input should be the same task")
		}
	}
	if err := c.ValidateInput(inputs, task); err != nil {
		return nil, err
	}
	for _, input := range inputs {
		output := structpb.Struct{}
		switch task {
		case callEndpoint:
			req := Request{}
			err := base.ConvertFromStructpb(input, &req)
			if err != nil {
				return nil, err
			}
			resp, err := req.sendReq()
			if err != nil {
				return nil, err
			}
			output, err = responseToStruct(resp)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not supported task: %s", task)
		}
		outputs = append(outputs, &output)
	}
	if err := c.ValidateOutput(outputs, task); err != nil {
		return nil, err
	}
	return outputs, nil
}

func responseToStruct(res Response) (structpb.Struct, error) {
	s := structpb.Struct{}
	outputJSON, err := json.Marshal(res)
	if err != nil {
		return s, err
	}
	err = protojson.Unmarshal(outputJSON, &s)
	if err != nil {
		return s, err
	}
	//handle JSON body
	if json.Valid([]byte(res.ResponseBody)) {
		var jsonBody any
		err = json.Unmarshal([]byte(res.ResponseBody), &jsonBody)
		if err == nil {
			str := structpb.Struct{}
			err = protojson.Unmarshal([]byte(res.ResponseBody), &str)
			s.Fields["response_body"] = structpb.NewStructValue(&str)
		}
	}
	return s, err
}

func (c *Operation) Test() (connectorPB.ConnectorResource_State, error) {
	return connectorPB.ConnectorResource_STATE_CONNECTED, nil
}
