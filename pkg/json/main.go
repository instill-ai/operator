package json

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

	om "github.com/instill-ai/component/pkg/objectmapper"
	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
)

const (
	getValue = "TASK_GET_VALUE"
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

type GetValueInput struct {
	Path       string `json:"path"`
	JSONString string `json:"json_string"`
}

type GetValueRes struct {
	Result any `json:"result"`
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
		case getValue:
			getValueStruct := GetValueInput{}
			err := base.ConvertFromStructpb(input, &getValueStruct)
			if err != nil {
				return nil, err
			}
			var obj map[string]interface{}
			err = json.Unmarshal([]byte(getValueStruct.JSONString), &obj)
			if err != nil {
				return nil, err
			}
			res, err := om.GetSrcValueByTag(obj, getValueStruct.Path)
			if err != nil {
				return nil, err
			}
			outputJson, err := json.Marshal(GetValueRes{Result: res})
			if err != nil {
				return nil, err
			}
			err = protojson.Unmarshal(outputJson, &output)
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

func (c *Operation) Test() (connectorPB.ConnectorResource_State, error) {
	return connectorPB.ConnectorResource_STATE_CONNECTED, nil
}
