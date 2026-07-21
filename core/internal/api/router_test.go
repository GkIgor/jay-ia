package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/GkIgor/jay-ia/core/internal/storage"
	"github.com/GkIgor/jay-ia/sdk/ipc"
)

func TestRouter_DispatchEnvelope_Success(t *testing.T) {
	router := NewRouter()

	router.Register(ipc.MsgCreateChat, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		var chatReq ipc.CreateChatRequest
		if err := ipc.UnmarshalPayload(req.Payload, &chatReq); err != nil {
			return nil, err
		}
		respPayload := ipc.CreateChatResponse{
			Chat: ipc.ChatDTO{ID: "chat-new", Title: chatReq.Title, Status: ipc.ChatActive},
		}
		return ipc.NewResponseEnvelope(req.RequestID, req.Type, respPayload)
	})

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateChat, "client-cpp", ipc.CreateChatRequest{Title: "Novo Chat"})
	rawBytes, _ := json.Marshal(reqEnv)

	respEnv := router.DispatchEnvelope(context.Background(), rawBytes)
	if respEnv == nil {
		t.Fatalf("esperava ResponseEnvelope não-nulo")
	}
	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava Status ErrSuccess (0), obteve %d", respEnv.Status)
	}
	if respEnv.RequestID != reqEnv.RequestID {
		t.Fatalf("esperava RequestID %s, obteve %s", reqEnv.RequestID, respEnv.RequestID)
	}

	var chatResp ipc.CreateChatResponse
	if err := ipc.UnmarshalPayload(respEnv.Payload, &chatResp); err != nil {
		t.Fatalf("falha ao desserializar resposta: %v", err)
	}
	if chatResp.Chat.Title != "Novo Chat" {
		t.Errorf("esperava Title 'Novo Chat', obteve %s", chatResp.Chat.Title)
	}
}

func TestRouter_Dispatch_JSONBytes(t *testing.T) {
	router := NewRouter()

	router.Register(ipc.MsgGetRegistration, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		return ipc.NewResponseEnvelope(req.RequestID, req.Type, ipc.RegisterClientResponse{
			Registration: ipc.RegistrationDTO{ID: "reg-1", Status: 1},
		})
	})

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetRegistration, "cli", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	outBytes := router.Dispatch(context.Background(), rawBytes)
	if len(outBytes) == 0 {
		t.Fatalf("esperava bytes de resposta não-vazios")
	}

	var respEnv ipc.ResponseEnvelope
	if err := json.Unmarshal(outBytes, &respEnv); err != nil {
		t.Fatalf("falha ao desserializar JSON de resposta: %v", err)
	}
	if respEnv.Status != ipc.ErrSuccess {
		t.Errorf("esperava Status 0, obteve %d", respEnv.Status)
	}
}

func TestRouter_InvalidJSONFrame(t *testing.T) {
	router := NewRouter()

	respEnv := router.DispatchEnvelope(context.Background(), []byte("{bad_json"))
	if respEnv.Status != ipc.ErrInvalidFormat {
		t.Fatalf("esperava Status ErrInvalidFormat (4000) para JSON inválido, obteve %d", respEnv.Status)
	}
}

func TestRouter_UnsupportedProtocolVersion(t *testing.T) {
	router := NewRouter()

	badReq := ipc.RequestEnvelope{
		ProtocolVersion: 99,
		RequestID:       "req-1",
		ClientID:        "client-1",
		Type:            ipc.MsgCreateChat,
	}
	rawBytes, _ := json.Marshal(badReq)

	respEnv := router.DispatchEnvelope(context.Background(), rawBytes)
	if respEnv.Status != ipc.ErrInvalidFormat {
		t.Fatalf("esperava Status ErrInvalidFormat (4000) para versão não suportada, obteve %d", respEnv.Status)
	}
}

func TestRouter_UnregisteredCommand(t *testing.T) {
	router := NewRouter()

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateVoiceSession, "client-1", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	respEnv := router.DispatchEnvelope(context.Background(), rawBytes)
	if respEnv.Status != ipc.ErrNotImplemented {
		t.Fatalf("esperava Status ErrNotImplemented (5001) para comando não cadastrado, obteve %d", respEnv.Status)
	}
}

func TestRouter_HandlerPanicIsolation(t *testing.T) {
	router := NewRouter()

	router.Register(ipc.MsgCreateChat, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		panic("simulated handler crash")
	})

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateChat, "client-1", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	// Não deve dar panic nem derrubar o teste
	respEnv := router.DispatchEnvelope(context.Background(), rawBytes)
	if respEnv.Status != ipc.ErrInternalDatabase {
		t.Fatalf("esperava Status ErrInternalDatabase (5000) ao capturar panic, obteve %d", respEnv.Status)
	}
	if respEnv.Error == nil || respEnv.Error.Message == "" {
		t.Errorf("esperava ErrorInfo preenchido na resposta de erro de panic")
	}
}

func TestRouter_DomainErrorMapping(t *testing.T) {
	router := NewRouter()

	router.Register(ipc.MsgGetChat, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		// Handler retorna erro de domínio padrão em Go
		return nil, storage.ErrNotFound
	})

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetChat, "client-1", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	respEnv := router.DispatchEnvelope(context.Background(), rawBytes)
	if respEnv.Status != ipc.ErrNotFound {
		t.Fatalf("esperava Status ErrNotFound (4004) mapeado do erro de domínio, obteve %d", respEnv.Status)
	}
}

func TestRouter_OverwriteRegister(t *testing.T) {
	router := NewRouter()

	router.Register(ipc.MsgCreateChat, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		return nil, errors.New("handler 1")
	})

	// Sobrescreve com o handler 2
	router.Register(ipc.MsgCreateChat, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		return ipc.NewResponseEnvelope(req.RequestID, req.Type, ipc.CreateChatResponse{
			Chat: ipc.ChatDTO{Title: "Handler 2 Winner"},
		})
	})

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateChat, "client-1", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	respEnv := router.DispatchEnvelope(context.Background(), rawBytes)
	if respEnv.Status != ipc.ErrSuccess {
		t.Fatalf("esperava sucesso no handler sobrescrito, obteve status %d", respEnv.Status)
	}

	var chatResp ipc.CreateChatResponse
	_ = ipc.UnmarshalPayload(respEnv.Payload, &chatResp)
	if chatResp.Chat.Title != "Handler 2 Winner" {
		t.Fatalf("esperava execução do segundo handler, obteve title %s", chatResp.Chat.Title)
	}
}

func TestRouter_MiddlewareExecution(t *testing.T) {
	router := NewRouter()

	var order []string

	mw1 := func(next Handler) Handler {
		return func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
			order = append(order, "mw1_before")
			res, err := next(ctx, req)
			order = append(order, "mw1_after")
			return res, err
		}
	}

	mw2 := func(next Handler) Handler {
		return func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
			order = append(order, "mw2_before")
			res, err := next(ctx, req)
			order = append(order, "mw2_after")
			return res, err
		}
	}

	router.Use(mw1, mw2)

	router.Register(ipc.MsgCreateChat, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		order = append(order, "handler")
		return ipc.NewResponseEnvelope(req.RequestID, req.Type, nil)
	})

	reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgCreateChat, "client-1", nil)
	rawBytes, _ := json.Marshal(reqEnv)

	_ = router.DispatchEnvelope(context.Background(), rawBytes)

	expectedOrder := []string{"mw1_before", "mw2_before", "handler", "mw2_after", "mw1_after"}
	if len(order) != len(expectedOrder) {
		t.Fatalf("esperava %d passos no middleware, obteve %d: %v", len(expectedOrder), len(order), order)
	}

	for i, exp := range expectedOrder {
		if order[i] != exp {
			t.Errorf("passo %d incorreto. Esperava %s, obteve %s", i, exp, order[i])
		}
	}
}

func TestRouter_ConcurrentDispatch(t *testing.T) {
	router := NewRouter()

	router.Register(ipc.MsgGetMessages, func(ctx context.Context, req *ipc.RequestEnvelope) (*ipc.ResponseEnvelope, error) {
		return ipc.NewResponseEnvelope(req.RequestID, req.Type, ipc.GetMessagesResponse{
			ChatID: "chat-concurrent",
		})
	})

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			reqEnv, _ := ipc.NewRequestEnvelope(ipc.MsgGetMessages, fmt.Sprintf("client-%d", id), nil)
			rawBytes, _ := json.Marshal(reqEnv)

			outBytes := router.Dispatch(context.Background(), rawBytes)
			if len(outBytes) == 0 {
				t.Errorf("resposta vazia para goroutine %d", id)
			}
		}(i)
	}

	wg.Wait()
}
