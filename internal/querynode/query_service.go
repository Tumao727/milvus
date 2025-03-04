// Copyright (C) 2019-2020 Zilliz. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under the License.

package querynode

import "C"
import (
	"context"
	"strconv"

	"go.uber.org/zap"

	miniokv "github.com/milvus-io/milvus/internal/kv/minio"
	"github.com/milvus-io/milvus/internal/log"
	"github.com/milvus-io/milvus/internal/msgstream"
	"github.com/milvus-io/milvus/internal/storage"
)

type queryService struct {
	ctx    context.Context
	cancel context.CancelFunc

	historical *historical
	streaming  *streaming

	queryCollections map[UniqueID]*queryCollection

	factory msgstream.Factory

	localChunkManager  storage.ChunkManager
	remoteChunkManager storage.ChunkManager
	localCacheEnabled  bool
}

func newQueryService(ctx context.Context,
	historical *historical,
	streaming *streaming,
	factory msgstream.Factory) *queryService {

	queryServiceCtx, queryServiceCancel := context.WithCancel(ctx)

	//TODO godchen: change this to configuration
	path, err := Params.Load("localStorage.Path")
	if err != nil {
		path = "/tmp/milvus/data"
	}
	enabled, _ := Params.Load("localStorage.enabled")
	localCacheEnabled, _ := strconv.ParseBool(enabled)

	localChunkManager := storage.NewLocalChunkManager(path)

	option := &miniokv.Option{
		Address:           Params.MinioEndPoint,
		AccessKeyID:       Params.MinioAccessKeyID,
		SecretAccessKeyID: Params.MinioSecretAccessKey,
		UseSSL:            Params.MinioUseSSLStr,
		CreateBucket:      true,
		BucketName:        Params.MinioBucketName,
	}

	client, err := miniokv.NewMinIOKV(ctx, option)
	if err != nil {
		panic(err)
	}
	remoteChunkManager := storage.NewMinioChunkManager(client)

	return &queryService{
		ctx:    queryServiceCtx,
		cancel: queryServiceCancel,

		historical: historical,
		streaming:  streaming,

		queryCollections: make(map[UniqueID]*queryCollection),

		factory: factory,

		localChunkManager:  localChunkManager,
		remoteChunkManager: remoteChunkManager,
		localCacheEnabled:  localCacheEnabled,
	}
}

func (q *queryService) close() {
	log.Debug("search service closed")
	for collectionID := range q.queryCollections {
		q.stopQueryCollection(collectionID)
	}
	q.queryCollections = make(map[UniqueID]*queryCollection)
	q.cancel()
}

func (q *queryService) addQueryCollection(collectionID UniqueID) {
	if _, ok := q.queryCollections[collectionID]; ok {
		log.Warn("query collection already exists", zap.Any("collectionID", collectionID))
		return
	}
	ctx1, cancel := context.WithCancel(q.ctx)
	qc := newQueryCollection(ctx1,
		cancel,
		collectionID,
		q.historical,
		q.streaming,
		q.factory,
		q.localChunkManager,
		q.remoteChunkManager,
		q.localCacheEnabled,
	)
	q.queryCollections[collectionID] = qc
}

func (q *queryService) hasQueryCollection(collectionID UniqueID) bool {
	_, ok := q.queryCollections[collectionID]
	return ok
}

func (q *queryService) stopQueryCollection(collectionID UniqueID) {
	sc, ok := q.queryCollections[collectionID]
	if !ok {
		log.Warn("stopQueryCollection failed, collection doesn't exist", zap.Int64("collectionID", collectionID))
		return
	}
	sc.close()
	sc.cancel()
	delete(q.queryCollections, collectionID)
}
