package server

import (
	"context"
	"net"
	"os"
	"testing"

	api "github.com/paulja/go-log/api/v1"
	"github.com/paulja/go-log/internal/auth"
	"github.com/paulja/go-log/internal/log"
	"github.com/paulja/go-log/tls"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Work directory: %v", pwd)
	for _, s := range []struct {
		scenario string
		user     string
		fn       func(
			t *testing.T,
			client api.LogClient,
			config *Config,
		)
	}{
		{
			scenario: "produce/consume a message to/from the log succeeds",
			user:     "root", fn: testProduceConsume,
		},
		{
			scenario: "produce/consume stream succeeds",
			user:     "root", fn: testProduceConsumeStream,
		},
		{
			scenario: "consume past log boundary fails",
			user:     "root", fn: testConsumePastBoundary,
		},
		{
			scenario: "unauthorised fails",
			user:     "nobody", fn: testUnauthorised,
		},
	} {
		t.Run(s.scenario, func(t *testing.T) {
			client, config, teardown := setupTest(t, s.user, nil)
			defer teardown()
			s.fn(t, client, config)
		})
	}
}

func setupTest(t *testing.T, user string, fn func(*Config)) (
	client api.LogClient,
	cfg *Config,
	teardown func(),
) {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	t.Logf("Run As User: %s", user)
	tls, err := tls.ClientConfig(user)
	require.NoError(t, err)
	clientOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tls)),
	}
	cc, err := grpc.NewClient(l.Addr().String(), clientOptions...)
	require.NoError(t, err)

	dir, err := os.MkdirTemp("", "server-test")
	require.NoError(t, err)

	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)

	authoriser := auth.NewDefault()
	cfg = &Config{CommitLog: clog, Authoriser: authoriser}
	if fn != nil {
		fn(cfg)
	}
	server, err := NewGRPCServer(cfg)
	require.NoError(t, err)

	go func() {
		server.Serve(l)
	}()

	client = api.NewLogClient(cc)

	return client, cfg, func() {
		server.Stop()
		cc.Close()
		l.Close()
		clog.Remove()
	}
}

func testProduceConsume(
	t *testing.T,
	client api.LogClient,
	config *Config,
) {
	ctx := context.Background()

	want := &api.Record{
		Value: []byte("hello world"),
	}

	produce, err := client.Produce(
		ctx,
		&api.ProduceRequest{
			Record: want,
		},
	)
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset,
	})
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)
}

func testConsumePastBoundary(
	t *testing.T,
	client api.LogClient,
	config *Config,
) {
	ctx := context.Background()

	produce, err := client.Produce(ctx, &api.ProduceRequest{
		Record: &api.Record{
			Value: []byte("hello world"),
		},
	})
	require.NoError(t, err)

	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: produce.Offset + 1,
	})
	if consume != nil {
		t.Fatal("consume not nil")
	}
	got := status.Code(err)
	want := status.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())
	if got != want {
		t.Fatalf("got err: %v, want: %v", got, want)
	}
}

func testProduceConsumeStream(
	t *testing.T,
	client api.LogClient,
	config *Config,
) {
	ctx := context.Background()

	records := []*api.Record{{
		Value:  []byte("firat message"),
		Offset: 0,
	}, {
		Value:  []byte("second message"),
		Offset: 1,
	}}

	{
		stream, err := client.ProduceStream(ctx)
		require.NoError(t, err)

		for offset, record := range records {
			err = stream.Send(&api.ProduceRequest{
				Record: record,
			})
			require.NoError(t, err)
			res, err := stream.Recv()
			require.NoError(t, err)
			if res.Offset != uint64(offset) {
				t.Fatalf("got offset: %d, want: %d", res.Offset, offset)
			}
		}
	}

	{
		stream, err := client.ConsumeStream(
			ctx,
			&api.ConsumeRequest{Offset: 0},
		)
		require.NoError(t, err)

		for i, record := range records {
			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, res.Record, &api.Record{
				Value:  record.Value,
				Offset: uint64(i),
			})
		}
	}
}

func testUnauthorised(
	t *testing.T,
	client api.LogClient,
	config *Config,
) {
	ctx := context.Background()
	produce, err := client.Produce(ctx, &api.ProduceRequest{
		Record: &api.Record{
			Value: []byte("hello world"),
		},
	})
	if produce != nil {
		t.Fatalf("produce response should be nil")
	}
	gotCode, wantCode := status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}
	consume, err := client.Consume(ctx, &api.ConsumeRequest{
		Offset: 0,
	})
	if consume != nil {
		t.Fatalf("consume response should be nil")
	}
	gotCode, wantCode = status.Code(err), codes.PermissionDenied
	if gotCode != wantCode {
		t.Fatalf("got code: %d, want: %d", gotCode, wantCode)
	}
}
