package base64

import (
	_ "embed"
	"encoding/base64"
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
	operatorName = "base64"
	encode       = "TASK_ENCODE"
	decode       = "TASK_DECODE"
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
	base.BaseOperation
	operator *Operator
}

type Base64 struct {
	Data string `json:"data"`
}

func Init(logger *zap.Logger, options OperatorOptions) base.IOperator {
	once.Do(func() {
		loader := configLoader.InitJSONSchema(logger)
		connDefs, err := loader.Load(operatorName, connectorPB.ConnectorType_CONNECTOR_TYPE_UNSPECIFIED, definitionJSON)
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

func (o *Operator) CreateOperation(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IOperation, error) {
	def, err := o.GetOperatorDefinitionByUid(defUid)
	if err != nil {
		return nil, err
	}
	return &Operation{
		BaseOperation: base.BaseOperation{
			Logger: logger, DefUid: defUid,
			Config:     config,
			Definition: def,
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
		base64Struct := Base64{}
		err := base.ConvertFromStructpb(input, &base64Struct)
		if err != nil {
			return nil, err
		}
		switch task {
		case encode:
			base64Struct.Data = Encode(base64Struct.Data)
		case decode:
			base64Struct.Data, err = Decode(base64Struct.Data)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not supported task: %s", task)
		}
		outputJson, err := json.Marshal(base64Struct)
		if err != nil {
			return nil, err
		}
		output := structpb.Struct{}
		err = protojson.Unmarshal(outputJson, &output)
		if err != nil {
			return nil, err
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

func Encode(str string) string {
	_, err := base64.StdEncoding.DecodeString(str)
	if err == nil {
		//already encoded
		return str
	}
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Decode(str string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return str, err
	}
	return string(b), nil
}
