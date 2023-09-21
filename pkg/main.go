package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
)

var once sync.Once
var operator base.IOperator

type Operator struct {
	base.BaseOperator
}

type OperatorOptions struct {
}

func Init(logger *zap.Logger, options OperatorOptions) base.IOperator {
	once.Do(func() {
		operator = &Operator{
			BaseOperator: base.BaseOperator{Logger: logger},
		}
	})
	return operator
}

func (c *Operator) CreateOperation(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IOperation, error) {
	return nil, fmt.Errorf("no operator uid: %s", defUid)
}
