package hello

import (
	"context"
	"fmt"
)

type Hello struct{}

type WorldRequest struct {
	Name string
}
type WorldResponse struct {
	Response string
}

func (h *Hello) Init(ctx context.Context) {}

func (h *Hello) Invoke(ctx context.Context, req *WorldRequest) (*WorldResponse, error) {
	return h.World(ctx, req)
}

func (h *Hello) World(ctx context.Context, req *WorldRequest) (*WorldResponse, error) {
	if req == nil {
		return nil, nil
	}
	rsp := WorldResponse{Response: "Hello, " + req.Name}
	return &rsp, nil
}

// this will panic if req is nil
func (h *Hello) Panic(ctx context.Context, req *WorldRequest) (*WorldResponse, error) {
	rsp := WorldResponse{Response: "Hello, " + req.Name}
	return &rsp, nil
}

func (h *Hello) Error(ctx context.Context) (*WorldResponse, error) {
	return nil, fmt.Errorf("method call failed")
}

// func (h *Hello) LogContext(ctx context.Context) error {
// 	for _, e := range os.Environ() {
// 		log.Printf("environment: %v", e)
// 	}

// 	mc, ok := mantil.FromContext(ctx)
// 	if !ok {
// 		return fmt.Errorf("context not found")
// 	}
// 	buf, err := json.Marshal(mc)
// 	if err != nil {
// 		return err
// 	}
// 	log.Printf("context: %s", buf)
// 	return nil
// }

func New() *Hello {
	return &Hello{}
}
