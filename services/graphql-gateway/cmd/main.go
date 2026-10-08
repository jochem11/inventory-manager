package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/auth"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/clients"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/gqlerr"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/user"
	authpb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/auth"
	userpb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/user"
	"github.com/jochem11/inventory-manager/shared/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := telemetry.Setup(ctx, "graphql-gateway")
	if err != nil {
		log.Fatalf("set up tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownTracing(ctx)
	}()

	// One connection per service the gateway calls.
	userConn := dial("USER_SERVICE_ADDR", "localhost:50051")
	defer userConn.Close()
	authConn := dial("AUTH_SERVICE_ADDR", "localhost:50052")
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	resolver := &graph.Resolver{
		UserResolver: user.NewResolver(userpb.NewUserServiceClient(userConn)),
		AuthResolver: auth.NewResolver(authClient),
	}
	// Checks the access token of every request with the auth-service's keys.
	authenticate := auth.Middleware(auth.NewVerifier(authClient), env("COOKIE_SECURE", "false") == "true")
	// The web app runs on another origin (port) and sends the refresh cookie.
	allowedOrigins := strings.Split(env("ALLOWED_ORIGINS", "http://localhost:3000"), ",")

	mux := http.NewServeMux()
	// A FORBIDDEN error makes the response a 403.
	mux.Handle("POST /graphql", gqlerr.HTTPStatus(graph.NewServer(resolver)))
	// The playground (GraphiQL) shares the endpoint: browsers GET it, queries POST.
	mux.Handle("GET /graphql", playground.Handler("Inventory manager", "/graphql"))
	mux.Handle("GET /{$}", http.RedirectHandler("/graphql", http.StatusFound))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              env("HTTP_ADDR", ":4000"),
		Handler:           traceRequests(cors(allowedOrigins, authenticate(mux))),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("GraphQL gateway listening on %s, playground at /graphql", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}

// traceRequests starts a span per HTTP request, continuing the caller's trace
// when it sends a traceparent header. Health checks are left out.
func traceRequests(h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, "graphql-gateway",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string { return r.Method + " " + r.URL.Path }),
		otelhttp.WithFilter(func(r *http.Request) bool { return r.URL.Path != "/healthz" }),
	)
}

// cors lets the web app call the gateway from its own origin, with
// credentials, so the browser sends and stores the refresh cookie. Other
// origins get no CORS headers, so browsers block their requests.
func cors(allowedOrigins []string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(allowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

// dial connects to the service whose address is in the env var addrKey.
func dial(addrKey, fallback string) *grpc.ClientConn {
	conn, err := clients.Dial(env(addrKey, fallback))
	if err != nil {
		log.Fatalf("create gRPC client for %s: %v", addrKey, err)
	}
	return conn
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
