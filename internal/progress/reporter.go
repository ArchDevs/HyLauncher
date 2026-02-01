package progress

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Stage string

const (
	StageIdle      Stage = "idle"
	StageVerify    Stage = "verify"
	StageJRE       Stage = "jre"
	StageButler    Stage = "butler"
	StagePWR       Stage = "pwr"
	StagePatch     Stage = "patch"
	StageOnlineFix Stage = "online-fix"
	StageLaunch    Stage = "launch"
	StageUpdate    Stage = "update"
	StageComplete  Stage = "complete"
)

// Data represents the progress data sent to frontend
type Data struct {
	Stage       Stage   `json:"stage"`
	Progress    float64 `json:"progress"`
	Message     string  `json:"message"`
	CurrentFile string  `json:"currentFile"`
	Speed       string  `json:"speed"`
	Downloaded  int64   `json:"downloaded"`
	Total       int64   `json:"total"`
}

type Reporter struct {
	ctx context.Context
}

func New(ctx context.Context) *Reporter {
	return &Reporter{ctx: ctx}
}

// Report sends a progress update to the frontend
func (p *Reporter) Report(stage Stage, progress float64, message string) {
	if p == nil || p.ctx == nil {
		return
	}

	runtime.EventsEmit(p.ctx, "progress-update", Data{
		Stage:    stage,
		Progress: progress,
		Message:  message,
	})
}

// ReportWithFile sends a progress update with file information
func (p *Reporter) ReportWithFile(stage Stage, progress float64, message string, currentFile string) {
	if p == nil || p.ctx == nil {
		return
	}

	runtime.EventsEmit(p.ctx, "progress-update", Data{
		Stage:       stage,
		Progress:    progress,
		Message:     message,
		CurrentFile: currentFile,
	})
}

func (p *Reporter) Reset() {
	if p == nil || p.ctx == nil {
		return
	}

	runtime.EventsEmit(p.ctx, "progress-update", Data{
		Stage:    StageIdle,
		Progress: 0,
		Message:  "",
	})
}

// ReportDownload sends a progress update with download metrics
func (p *Reporter) ReportDownload(stage Stage, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
	if p == nil || p.ctx == nil {
		return
	}

	runtime.EventsEmit(p.ctx, "progress-update", Data{
		Stage:       stage,
		Progress:    progress,
		Message:     message,
		CurrentFile: currentFile,
		Speed:       speed,
		Downloaded:  downloaded,
		Total:       total,
	})
}

// Scaler wraps a Reporter to scale progress within a range
type Scaler struct {
	reporter *Reporter
	stage    Stage
	start    float64
	end      float64
}

// NewScaler creates a progress scaler for a specific range
func NewScaler(reporter *Reporter, stage Stage, start, end float64) *Scaler {
	return &Scaler{
		reporter: reporter,
		stage:    stage,
		start:    start,
		end:      end,
	}
}

// scale converts a progress value (0-100) to the scaled range
func (s *Scaler) scale(progress float64) float64 {
	return s.start + (progress * (s.end - s.start) / 100.0)
}

// Report for the scaler - maps progress to the scaled range
func (s *Scaler) Report(stage Stage, progress float64, message string) {
	// Use the scaler's stage if provided, otherwise use the passed stage
	actualStage := s.stage
	if actualStage == "" {
		actualStage = stage
	}
	s.reporter.Report(actualStage, s.scale(progress), message)
}

// ReportWithFile for the scaler
func (s *Scaler) ReportWithFile(stage Stage, progress float64, message string, currentFile string) {
	actualStage := s.stage
	if actualStage == "" {
		actualStage = stage
	}
	s.reporter.ReportWithFile(actualStage, s.scale(progress), message, currentFile)
}

// ReportDownload for the scaler - includes download metrics
func (s *Scaler) ReportDownload(stage Stage, progress float64, message string, currentFile string, speed string, downloaded, total int64) {
	actualStage := s.stage
	if actualStage == "" {
		actualStage = stage
	}
	s.reporter.ReportDownload(actualStage, s.scale(progress), message, currentFile, speed, downloaded, total)
}
