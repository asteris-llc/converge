// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"encoding/json"

	"google.golang.org/grpc/metadata"

	"github.com/asteris-llc/converge/apply"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/healthcheck"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

type executor struct{}

type statusResponseStream interface {
	Send(*pb.StatusResponse) error
	SendHeader(metadata.MD) error
}

func (e *executor) edgeMeta(ctx context.Context, g *graph.Graph) (metadata.MD, error) {
	logger := getLogger(ctx).WithField("function", "executor.edgeMeta")

	edges, err := json.Marshal(g.Edges())
	if err != nil {
		logger.WithError(err).Error("could not serialize edges")
		return nil, errors.Wrapf(err, "serializing edges")
	}

	return metadata.New(map[string]string{"edges": string(edges)}), nil
}

func (e *executor) sendMeta(ctx context.Context, g *graph.Graph, stream statusResponseStream) error {
	logger := getLogger(ctx).WithField("function", "executor.sendMeta")

	// dehydrate graph edges and send them in the header metadata
	meta, err := e.edgeMeta(ctx, g)
	if err != nil {
		// already logged, don't log here
		return errors.Wrap(err, "preparing metadata")
	}

	if err = stream.SendHeader(meta); err != nil {
		logger.WithError(err).Error("could not send metadata")
		return errors.Wrap(err, "sending metadata")
	}

	return nil
}

func (e *executor) stageNotifier(stage pb.StatusResponse_Stage, stream statusResponseStream) *graph.Notifier {
	return &graph.Notifier{
		Pre: func(meta *node.Node) error {
			return stream.Send(&pb.StatusResponse{
				Id:    meta.ID, // TODO: deprecated, remove in 0.4.0
				Stage: stage,
				Run:   pb.StatusResponse_STARTED,
				Meta:  pb.MetaFromNode(meta),
			})
		},
		Post: func(meta *node.Node) error {
			response := statusResponseFromPrintable(
				meta,
				meta.Value().(human.Printable),
				stage,
				pb.StatusResponse_FINISHED,
			)

			return stream.Send(response)
		},
	}
}

func (e *executor) sendPlan(ctx context.Context, stream statusResponseStream, in *graph.Graph) (*graph.Graph, error) {
	out, err := plan.WithNotify(ctx, in, e.stageNotifier(pb.StatusResponse_PLAN, stream))
	if err != nil && err != plan.ErrTreeContainsErrors {
		return nil, err
	}
	return out, nil
}

func (e *executor) Plan(in *pb.LoadRequest, stream pb.Executor_PlanServer) error {
	logger, ctx := setIDLogger(stream.Context())
	logger = logger.WithField("function", "executor.Plan")

	loaded, err := in.Load(ctx)
	if err != nil {
		return err
	}

	if err = e.sendMeta(ctx, loaded, stream); err != nil {
		return err
	}

	// send the plan
	_, err = e.sendPlan(ctx, stream, loaded)
	if err != nil {
		logger.WithError(err).WithField("location", in.Location).Error("planning failed")
		return errors.Wrapf(err, "planning %s", in.Location)
	}

	return nil
}

func (e *executor) sendHealthCheck(ctx context.Context, stream statusResponseStream, in *graph.Graph) (*graph.Graph, error) {
	out, err := healthcheck.WithNotify(ctx, in, e.stageNotifier(pb.StatusResponse_PLAN, stream))
	if err != nil && err != plan.ErrTreeContainsErrors {
		return nil, err
	}
	return out, nil
}

func (e *executor) HealthCheck(in *pb.LoadRequest, stream pb.Executor_HealthCheckServer) error {
	logger, ctx := setIDLogger(stream.Context())
	logger = logger.WithField("function", "executor.Plan")

	loaded, err := in.Load(ctx)
	if err != nil {
		return err
	}

	if err = e.sendMeta(ctx, loaded, stream); err != nil {
		return err
	}

	// send the plan
	planned, err := e.sendPlan(ctx, stream, loaded)
	if err != nil {
		logger.WithError(err).WithField("location", in.Location).Error("planning failed")
		return errors.Wrapf(err, "planning %s", in.Location)
	}

	_, err = e.sendHealthCheck(ctx, stream, planned)
	if err != nil {
		logger.WithError(err).WithField("location", in.Location).Error("health check failed")
		return errors.Wrapf(err, "health check %s", in.Location)
	}

	return nil
}

func (e *executor) sendApply(ctx context.Context, stream statusResponseStream, in *graph.Graph) (*graph.Graph, error) {
	out, err := apply.WithNotify(ctx, in, e.stageNotifier(pb.StatusResponse_APPLY, stream))
	if err != nil && err != apply.ErrTreeContainsErrors {
		return nil, err
	}
	return out, nil
}

func (e *executor) Apply(in *pb.LoadRequest, stream pb.Executor_ApplyServer) error {
	logger, ctx := setIDLogger(stream.Context())
	logger = logger.WithField("function", "executor.Apply")

	loaded, err := in.Load(ctx)
	if err != nil {
		return err
	}

	if err = e.sendMeta(ctx, loaded, stream); err != nil {
		return err
	}

	_, err = e.sendApply(ctx, stream, loaded)
	if err != nil {
		return errors.Wrapf(err, "applying %s", in.Location)
	}

	return nil
}
