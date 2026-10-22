// Single-process gRPC demo: the server and the client live in the same program.
//
// The transport is bufconn, an in-memory net.Listener shipped with grpc-go:
//   - no TCP port, no Unix socket, no file on disk — bytes move through memory
//   - nothing outside this process can connect to it
//   - it is still the full gRPC stack (HTTP/2 framing, protobuf, status codes,
//     interceptors, deadlines), so splitting into two processes later only
//     means swapping the listener and the dialer.
//
// Regenerate the types/stubs after editing proto/todo.proto:
//
//go:generate protoc -I proto --go_out=. --go_opt=module=localgrpc --go-grpc_out=. --go-grpc_opt=module=localgrpc proto/todo.proto
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "localgrpc/gen/todopb"
)

// ------------------------------------------------------------------ server side

type todoServer struct {
	pb.UnimplementedTodoServiceServer // forward-compatible with new RPCs

	mu     sync.Mutex // handlers run concurrently
	tasks  map[int32]*pb.Task
	nextID int32
}

func newTodoServer() *todoServer {
	return &todoServer{tasks: map[int32]*pb.Task{}, nextID: 1}
}

func (s *todoServer) AddTask(_ context.Context, req *pb.AddTaskRequest) (*pb.Task, error) {
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := &pb.Task{Id: s.nextID, Title: req.GetTitle()}
	s.tasks[t.Id] = t
	s.nextID++
	return t, nil
}

func (s *todoServer) Complete(_ context.Context, req *pb.CompleteRequest) (*pb.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[req.GetId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no task with id %d", req.GetId())
	}
	t.Done = true
	return t, nil
}

func (s *todoServer) selectTasks(onlyOpen bool) []*pb.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*pb.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if !(onlyOpen && t.Done) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}

func (s *todoServer) List(_ context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	return &pb.ListResponse{Tasks: s.selectTasks(req.GetOnlyOpen())}, nil
}

func (s *todoServer) Watch(req *pb.ListRequest, stream pb.TodoService_WatchServer) error {
	for _, t := range s.selectTasks(req.GetOnlyOpen()) {
		if err := stream.Send(t); err != nil {
			return err
		}
	}
	return nil
}

// startServer runs the gRPC server on an in-memory listener and returns a
// client connection to it plus a cleanup func.
func startServer() (*grpc.ClientConn, func(), error) {
	lis := bufconn.Listen(1 << 20) // 1 MiB in-memory buffer

	srv := grpc.NewServer()
	pb.RegisterTodoServiceServer(srv, newTodoServer())
	go func() {
		if err := srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Printf("server stopped: %v", err)
		}
	}()

	// "passthrough:///" stops gRPC from trying DNS on the dummy target name;
	// the dialer hands it the in-memory pipe instead of a real socket.
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // nothing leaves the process
	)
	if err != nil {
		srv.Stop()
		return nil, nil, err
	}
	cleanup := func() {
		conn.Close()
		srv.GracefulStop()
	}
	return conn, cleanup, nil
}

// ------------------------------------------------------------------ client side

func runClient(ctx context.Context, client pb.TodoServiceClient) error {
	for _, title := range []string{"Write proto", "Generate stubs", "Ship it"} {
		t, err := client.AddTask(ctx, &pb.AddTaskRequest{Title: title})
		if err != nil {
			return err
		}
		fmt.Printf("added    #%d %s\n", t.Id, t.Title)
	}

	done, err := client.Complete(ctx, &pb.CompleteRequest{Id: 2})
	if err != nil {
		return err
	}
	fmt.Printf("complete #%d %s -> done=%v\n", done.Id, done.Title, done.Done)

	fmt.Println("open tasks (unary List):")
	list, err := client.List(ctx, &pb.ListRequest{OnlyOpen: true})
	if err != nil {
		return err
	}
	for _, t := range list.Tasks {
		fmt.Printf("  #%d %s\n", t.Id, t.Title)
	}

	fmt.Println("all tasks (streaming Watch):")
	stream, err := client.Watch(ctx, &pb.ListRequest{})
	if err != nil {
		return err
	}
	for {
		t, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		mark := " "
		if t.Done {
			mark = "x"
		}
		fmt.Printf("  #%d %s [%s]\n", t.Id, t.Title, mark)
	}

	// Errors travel as real gRPC status codes, same as over a network.
	_, err = client.Complete(ctx, &pb.CompleteRequest{Id: 99})
	if st, ok := status.FromError(err); ok && err != nil {
		fmt.Printf("error    %s: %s\n", st.Code(), st.Message())
	}
	return nil
}

func main() {
	conn, cleanup, err := startServer()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := runClient(ctx, pb.NewTodoServiceClient(conn)); err != nil {
		log.Fatal(err)
	}
}
