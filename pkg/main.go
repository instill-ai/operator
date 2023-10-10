package pkg

import (
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/operator/pkg/base64"
	"github.com/instill-ai/operator/pkg/downloadurl"
	"github.com/instill-ai/operator/pkg/end"
	"github.com/instill-ai/operator/pkg/json"
	"github.com/instill-ai/operator/pkg/rest"
	"github.com/instill-ai/operator/pkg/start"
	"github.com/instill-ai/operator/pkg/textextraction"
)

var (
	once     sync.Once
	operator base.IOperator
)

type Operator struct {
	base.Operator
	operatorUIDMap map[uuid.UUID]base.IOperator
}

func Init(logger *zap.Logger) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			Operator:       base.Operator{Component: base.Component{Logger: logger}},
			operatorUIDMap: map[uuid.UUID]base.IOperator{},
		}
		operator.(*Operator).ImportDefinitions(base64.Init(logger))
		operator.(*Operator).ImportDefinitions(textextraction.Init(logger))
		operator.(*Operator).ImportDefinitions(start.Init(logger))
		operator.(*Operator).ImportDefinitions(end.Init(logger))
		operator.(*Operator).ImportDefinitions(json.Init(logger))
		operator.(*Operator).ImportDefinitions(rest.Init(logger))
		operator.(*Operator).ImportDefinitions(downloadurl.Init(logger))

	})
	return operator
}

func (o *Operator) GetOperatorUIDMap() map[uuid.UUID]base.IOperator {
	return o.operatorUIDMap
}

func (o *Operator) ImportDefinitions(op base.IOperator) {
	for _, v := range op.ListOperatorDefinitions() {
		err := o.AddOperatorDefinition(v)
		if err != nil {
			panic(err)
		}
		o.operatorUIDMap[uuid.FromStringOrNil(v.Uid)] = op
	}
}

func (o *Operator) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	return o.operatorUIDMap[defUID].CreateExecution(defUID, task, config, logger)
}
