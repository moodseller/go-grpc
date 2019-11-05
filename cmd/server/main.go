package main

import (
	"crypto/tls"
	"do/pkg/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"do/internal/services/user"
	"do/internal/version"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	logMiddleware "do/internal/middleware/log"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	metrics     = prometheus.NewRegistry()
	grpcMetrics = grpcprometheus.NewServerMetrics()

	keepalivePolicy = keepalive.EnforcementPolicy{
		MinTime:             10 * time.Second, // If a client pings more than once every x duration, terminate the connection.
		PermitWithoutStream: true,             // Allow pings even when there are no active streams
	}

	keepaliveParams = keepalive.ServerParameters{
		MaxConnectionIdle:     20 * time.Minute,  // If a client is idle for given duration, send a GOAWAY.
		MaxConnectionAge:      60 * time.Minute,  // If any connection is alive for more than given duration, send a GOAWAY.
		MaxConnectionAgeGrace: 10 * time.Second,  // Allow given duration for pending RPCs to complete before forcibly closing connections
		Time:                  120 * time.Second, // Ping the client if it is idle for given duration to ensure the connection is still active.
		Timeout:               30 * time.Second,  // Wait given duration for the ping ack before assuming the connection is dead.
	}

	// graceShutdownPeriod is a time to wait for grpc server to shutdown.
	graceShutdownPeriod = 10 * time.Second

	// conf holds global app configuration.
	conf *config
)

const (
	grpcPort    = ":10000"
	metricsPort = ":10002"
)

type server struct {
	instance     *grpc.Server
	onShutdownCb func()
}

func main() {
	grpcMetrics.EnableHandlingTimeHistogram()
	metrics.MustRegister(grpcMetrics)

	grpcServer := newServer(
		withCredentials(),
		grpc.KeepaliveEnforcementPolicy(keepalivePolicy),
		grpc.KeepaliveParams(keepaliveParams),
		grpcmiddleware.WithUnaryServerChain(
			grpcMetrics.UnaryServerInterceptor(),
			logMiddleware.UnaryServerInterceptor(slog.Logger(), unaryLogSkipper),
			grpcrecovery.UnaryServerInterceptor(),
		),
		grpcmiddleware.WithStreamServerChain(
			grpcMetrics.StreamServerInterceptor(),
			logMiddleware.StreamServerInterceptor(slog.Logger()),
			grpcrecovery.StreamServerInterceptor(),
		),
	)

	grpcServer.registerServices(func(s *grpc.Server) {
		user.RegisterServer(s)
	})

	grpcMetrics.InitializeMetrics(grpcServer.instance)
	go startMetrics()

	grpcServer.onShutdown(func() {
		// Initiate database close connections and such
	})

	grpcServer.start(conf.isDevelopment())
}

func init() {
	conf = &config{}
	if err := conf.parse(); err != nil {
		slog.Fatalf("cannot parse env config: %v, version: %s", err, version.VERSION)
	}

	slog.SetFormatter(conf.isDevelopment())
	slog.SetLevel(logrus.InfoLevel)
	logMiddleware.ReplaceGrpcLogger(slog.Copy())
}

func newServer(opt ...grpc.ServerOption) *server {
	instance := grpc.NewServer(opt...)
	return &server{instance: instance}
}

func (s *server) registerServices(registerFunc func(s *grpc.Server)) {
	registerFunc(s.instance)
	reflection.Register(s.instance)
}

func (s *server) onShutdown(cb func()) {
	s.onShutdownCb = cb
}

func (s *server) start(isDevelopment bool) {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		slog.Fatalf("failed to listen: %v", err)
	}

	// Handle graceful shutdown.
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		<-gracefulStop

		// Try to do graceful stop to signal all active connections to go away.
		// Since we have grpc bidi streams in most cases GracefulStop will
		// never finish so we need to force stop after gracePeriod.
		go func() {
			s.instance.GracefulStop()
		}()

		s.onShutdownCb()
		// Do not sleep on development environment.
		if !isDevelopment {
			time.Sleep(graceShutdownPeriod)
		}
		s.instance.Stop()
		os.Exit(0)
	}()

	slog.Infof("server ready on port %s", grpcPort)
	if err := s.instance.Serve(lis); err != nil {
		slog.Fatalf("failed to serve: %v", err)
	}
}

func unaryLogSkipper(method string) bool {
	// if method == healthCheckMethod {
	// 	return true
	// }
	return false
}

func withCredentials() grpc.ServerOption {
	publicKey := os.Getenv("CERT_PUBLIC_KEY")
	privateKey := os.Getenv("CERT_PRIVATE_KEY")
	if publicKey == "" || privateKey == "" {
		return grpc.Creds(nil)
	}

	cert, err := tls.X509KeyPair([]byte(publicKey), []byte(privateKey))
	if err != nil {
		slog.Fatalf("cannot load certs: %v", err)
	}
	creds := credentials.NewServerTLSFromCert(&cert)
	return grpc.Creds(creds)
}

func startMetrics() {
	srv := &http.Server{Handler: promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}), Addr: metricsPort}
	if err := srv.ListenAndServe(); err != nil {
		slog.Fatalf("cannot start prometheus monitoring server: %v", err)
	}
}
