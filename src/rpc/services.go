package rpc

import (
	"context"

	"github.com/migalabs/armiarma/src/base"
)

type MetadataArgs struct {
	Data []byte
}
type MetadataReply struct {
	Data []byte
}
type MetadataService struct{}

func (t *MetadataService) MetaDataRPCv1(ctx context.Context, argType MetadataArgs, replyType *MetadataReply) error {
	localLogger := base.CreateLogger(base.CreateLogOpts("METADATA", "temrinal", "text", "debug"))

	localLogger.Infof("Received a Metadata call")
	replyType.Data = argType.Data
	return nil
}
