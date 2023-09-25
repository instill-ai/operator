package pipeline

import (
	_ "embed"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/component/pkg/configLoader"
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
	panic("we should no use this function")
}

func (c *Operation) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	panic("we should no use this function")
}
