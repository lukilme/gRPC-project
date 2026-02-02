package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"ifpb.com/microservices/shipping/config"
	"ifpb.com/microservices/shipping/internal/adapters/db"
	"ifpb.com/microservices/shipping/internal/adapters/grpc"
	"ifpb.com/microservices/shipping/internal/application/core/api"
)

func init() {
	log.SetFormatter(customLogger{
		formatter: log.JSONFormatter{FieldMap: log.FieldMap{
			"msg": "message",
		}},
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

type customLogger struct {
	formatter log.JSONFormatter
}

func (l customLogger) Format(entry *log.Entry) ([]byte, error) {
	span := trace.SpanFromContext(entry.Context)
	entry.Data["trace_id"] = span.SpanContext().TraceID().String()
	entry.Data["span_id"] = span.SpanContext().SpanID().String()
	entry.Data["Context"] = span.SpanContext()
	return l.formatter.Format(entry)
}

func main() {

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {

		log.Fatalf("Failed to connect to database. Error: %v", err)
	}

	application := api.NewApplication(dbAdapter)
	grpcAdapter := grpc.NewAdapter(application, config.GetApplicationPort())
	grpcAdapter.Run()
}
