package pkg

import (
	"fmt"
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/operator/pkg/base64"
	"github.com/instill-ai/operator/pkg/textextraction"
)

var (
	once     sync.Once
	operator base.IOperator
)

type Operator struct {
	base.BaseOperator
	base64Operator         base.IOperator
	textExtractionOperator base.IOperator
}

type OperatorOptions struct {
	Base64         base64.OperatorOptions
	TextExtraction textextraction.OperatorOptions
}

func Init(logger *zap.Logger, options OperatorOptions) base.IOperator {
	once.Do(func() {
		base64Operator := base64.Init(logger, options.Base64)
		textExtractionOperator := textextraction.Init(logger, options.TextExtraction)
		operator = &Operator{
			BaseOperator:           base.BaseOperator{Logger: logger},
			base64Operator:         base64Operator,
			textExtractionOperator: textExtractionOperator,
		}
		for _, uid := range base64Operator.ListOperatorDefinitionUids() {
			def, err := base64Operator.GetOperatorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = operator.AddOperatorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
		for _, uid := range textExtractionOperator.ListOperatorDefinitionUids() {
			def, err := textExtractionOperator.GetOperatorDefinitionByUid(uid)
			if err != nil {
				logger.Error(err.Error())
			}
			err = operator.AddOperatorDefinition(uid, def.GetId(), def)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
	})
	return operator
}

func (o *Operator) CreateOperation(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (base.IOperation, error) {
	switch {
	case o.base64Operator.HasUid(defUid):
		return o.base64Operator.CreateOperation(defUid, config, logger)
	case o.textExtractionOperator.HasUid(defUid):
		return o.textExtractionOperator.CreateOperation(defUid, config, logger)
	default:
		return nil, fmt.Errorf("no operator uid: %s", defUid)
	}
}
