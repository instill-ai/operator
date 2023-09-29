package textextraction

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
	"github.com/instill-ai/operator/pkg/base64"
)

const (
	taskExtractFromPath = "TASK_EXTRACT_FROM_PATH"
	taskExtractFromFile = "TASK_EXTRACT_FROM_FILE"
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

type FromPath struct {
	FilePath string `json:"path"`
}

type FromFile struct {
	FileContents string `json:"file_contents"`
	ContentType  string `json:"content_type"`
}

type Output struct {
	Text string `json:"text"`
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
		op := Output{}
		switch task {
		case taskExtractFromPath:
			obj := FromPath{}
			err := base.ConvertFromStructpb(input, &obj)
			if err != nil {
				return nil, err
			}
			op.Text, err = PathToText(obj.FilePath)
			if err != nil {
				return nil, err
			}
		case taskExtractFromFile:
			obj := FromFile{}
			err := base.ConvertFromStructpb(input, &obj)
			if err != nil {
				return nil, err
			}
			obj.FileContents, err = base64.Decode(obj.FileContents)
			if err != nil {
				return nil, err
			}
			op.Text, err = BytesToText([]byte(obj.FileContents), obj.ContentType)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("not supported task: %s", task)
		}
		outputJson, err := json.Marshal(op)
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
