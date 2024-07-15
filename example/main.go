package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"example/auth"
	"example/target"

	"github.com/justenwalker/mack/crypt/random"
	"github.com/justenwalker/mack/macaroon"
	"github.com/justenwalker/mack/macaroon/thirdparty"
	"github.com/justenwalker/mack/sensible"
)

func main() {
	// for the paranoid: ensures we didn't build this binary with -tags test_random
	random.MustNotBeTest()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		log.Fatal("ERROR:", err)
	}
}

const (
	authAPIAddr   = "127.0.0.1:8080"
	targetAPIAddr = "127.0.0.1:8081"

	authServiceLocation   = "http://127.0.0.1:8080"
	targetServiceLocation = "http://127.0.0.1:8081"
)

func run(ctx context.Context) error {
	// Just use the sensible scheme
	scheme := sensible.Scheme()

	srvCtx, srvCancel := context.WithCancel(ctx)
	doneCh := make(chan struct{})
	defer func() {
		log.Println("Shutting Down API Servers...")
		srvCancel()
		<-doneCh
		log.Println("Shutting Down API Servers... complete")
	}()

	log.Println("Starting API Servers")
	go func() {
		defer close(doneCh)
		if err := startServers(srvCtx, scheme); err != nil {
			log.Fatalf("ERROR: failed to start servers: %v", err)
		}
	}()

	// Creates a macaroon third-party for the auth service to discharge auth caveats
	// This logs into the auth service and acquires the access token.
	var thirdParties thirdparty.Set
	authThirdParty, err := auth.NewThirdParty(ctx, authServiceLocation, auth.Credentials{
		Username: "foo",
		Password: "secret",
	})
	if err != nil {
		return fmt.Errorf("auth.NewThirdParty: %w", err)
	}
	thirdParties = append(thirdParties, authThirdParty)

	// Create a Target Service API Client - the thing we want to call api methods on
	targetAPIClient, err := target.NewAPIClient(targetServiceLocation, authThirdParty.AccessToken())
	if err != nil {
		return fmt.Errorf("target.NewAPIClient: %w", err)
	}

	// Acquire a macaroon for the target service
	log.Println("- Acquire Macaroon")
	m, err := targetAPIClient.GetMacaroon(ctx, "myorg", "myapp")
	if err != nil {
		return fmt.Errorf("targetAPIClient.GetMacaroon: %w", err)
	}
	log.Println(" => Macaroon:", &m)

	// Discharge the macaroon using our thirdParties
	log.Println("- Discharge Macaroon Caveats")
	discharge, err := thirdParties.Discharge(ctx, &m)
	if err != nil {
		return fmt.Errorf("thirdPartySet.Discharge: %w", err)
	}
	for i := range discharge {
		log.Printf(" => Discharge[%d]: %s", i, &discharge[i])
	}

	// Prepare the macaroon stack for the api request
	log.Println("- Preparing Macaroon Stack")
	stack, err := scheme.PrepareStack(&m, discharge)
	if err != nil {
		return fmt.Errorf("scheme.PrepareStack: %w", err)
	}
	for i := range stack {
		log.Printf(" => Stack[%d]: %s", i, &stack[i])
	}

	// Executes API Calls to the Target Service using a macaroon token.
	// 1. This API should succeed, since it meets all the caveats.
	log.Println("1. Execute Successful Request: /myorg/myapp/do")
	res, err := targetAPIClient.DoOperation(ctx, stack, target.Operation{
		Org:       "myorg",
		App:       "myapp",
		Operation: "foo",
		Args:      []string{"bar"},
	})
	if err != nil {
		return fmt.Errorf("targetAPIClient.DoOperation: %w", err)
	}
	log.Println("Result:", res)

	// 2. This API call should fail, since it doesn't meet the requirements.
	log.Println("2. Execute Failing Request /otherorg/myapp/do")
	_, err = targetAPIClient.DoOperation(ctx, stack, target.Operation{
		Org:       "otherorg",
		App:       "myapp",
		Operation: "foo",
		Args:      []string{"bar"},
	})
	if err != nil {
		log.Println("API Error:", err)
	}

	// 3. This API call should fail, since we didn't bind the request
	log.Println("3. Execute Failing Request - Macaroon Verification Failure due to discharge not bound to target")
	_, err = targetAPIClient.DoOperation(ctx, macaroon.Stack{m, discharge[0]}, target.Operation{
		Org:       "myorg",
		App:       "myapp",
		Operation: "foo",
		Args:      []string{"bar"},
	})
	if err != nil {
		log.Println("API Error:", err)
	}
	return nil
}

func startServers(ctx context.Context, scheme *macaroon.Scheme) error {
	// Start listening
	authListener, err := net.Listen("tcp", authAPIAddr)
	if err != nil {
		return err
	}
	defer closeListener(authListener)
	targetListener, err := net.Listen("tcp", targetAPIAddr)
	if err != nil {
		return err
	}
	defer closeListener(targetListener)

	var wg sync.WaitGroup
	wg.Add(2)

	// Create auth api server and start it
	authAPI, err := auth.NewAPI(scheme, authServiceLocation)
	if err != nil {
		return fmt.Errorf("auth.NewAPI: %w", err)
	}
	authAPIServer := &http.Server{Handler: authAPI.Handler()}
	go func() {
		defer wg.Done()
		serveErr := authAPIServer.Serve(authListener)
		if errors.Is(serveErr, http.ErrServerClosed) {
			return
		}
		log.Fatalf("ERROR: targetAPIServer.Serve: %v", serveErr)
	}()
	go shutdownOnContextCancel(ctx, 5*time.Second, authAPIServer)

	// Create target api server and start it
	targetAPI, err := target.NewAPI(ctx, target.APIConfig{
		Scheme:      scheme,
		Location:    targetServiceLocation,
		SecretKey:   "secret-key",
		AuthService: authServiceLocation,
	})
	if err != nil {
		return fmt.Errorf("NewTargetService: %w", err)
	}
	targetAPIServer := &http.Server{
		Handler: targetAPI.Handler(),
	}
	go func() {
		defer wg.Done()
		serveErr := targetAPIServer.Serve(targetListener)
		if errors.Is(serveErr, http.ErrServerClosed) {
			return
		}
		log.Fatalf("ERROR: targetAPIServer.Serve: %v", serveErr)
	}()
	go shutdownOnContextCancel(ctx, 5*time.Second, targetAPIServer)
	wg.Wait()
	return nil
}

func shutdownOnContextCancel(ctx context.Context, timeout time.Duration, srv *http.Server) {
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	<-ctx.Done()
	_ = srv.Shutdown(shutdownCtx)
}

func closeListener(ln net.Listener) {
	_ = ln.Close()
}
