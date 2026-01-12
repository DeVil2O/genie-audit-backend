package logger

import (
	"context"
	"fmt"
	"genie-audit-backend/helpers"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/mattn/go-colorable"
)

// Global logger instance
var Log *Logger

type Logger struct {
	*logrus.Logger
}

// Keys to extract from context
const (
	TraceIDKey = "traceID"
	TxnIDKey   = "txnID"
)

// Initialize initializes the global logger instance.
func Initialize() {
	level := os.Getenv(helpers.LOG_LEVEL)
	logLevel := logrus.InfoLevel
	if level == "debug" {
		logLevel = logrus.DebugLevel
	}

	Log = &Logger{logrus.New()}
	Log.SetLevel(logLevel)

	Log.SetOutput(colorable.NewColorableStdout())
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})
}

func (l *Logger) InfofWithContext(ctx context.Context, format string, args ...interface{}) {
	l.Infof(format+GetTraceIDandTxnID(ctx), args...)
}

func (l *Logger) ErrorfWithContext(ctx context.Context, format string, args ...interface{}) {
	l.Errorf(format+GetTraceIDandTxnID(ctx), args...)
}

func (l *Logger) DebugfWithContext(ctx context.Context, format string, args ...interface{}) {
	l.Debugf(format+GetTraceIDandTxnID(ctx), args...)
}

func (l *Logger) WarnfWithContext(ctx context.Context, format string, args ...interface{}) {
	l.Warnf(format+GetTraceIDandTxnID(ctx), args...)
}

func (l *Logger) InfoWithContext(ctx context.Context, args ...interface{}) {
	args = append(args, GetTraceIDandTxnID(ctx))
	l.Info(args...)
}

func (l *Logger) ErrorWithContext(ctx context.Context, args ...interface{}) {
	args = append(args, GetTraceIDandTxnID(ctx))
	l.Error(args...)
}

func (l *Logger) DebugWithContext(ctx context.Context, args ...interface{}) {
	args = append(args, GetTraceIDandTxnID(ctx))
	l.Debug(args...)
}
func (l *Logger) WarnWithContext(ctx context.Context, args ...interface{}) {
	args = append(args, GetTraceIDandTxnID(ctx))
	l.Warn(args...)
}

func GetTraceIDandTxnID(ctx context.Context) string {
	// Extract traceID and txnID from context
	msg := ""
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		msg += fmt.Sprintf(" ,traceID: %s", traceID)
	}
	if txnID, ok := ctx.Value(TxnIDKey).(string); ok {
		msg += fmt.Sprintf(" ,txnID: %s", txnID)
	}
	return msg
}
