package main

import (
	"fmt"
	"strings"
	"time"
)

type ProgressPrinter struct {
	focusDuration  time.Duration
	breakDuration  time.Duration
	currentPrinted bool
	v              *PomConSimplePrinter
}

func (s *ProgressPrinter) VisitNode(node *PomConNode) {
	if !s.currentPrinted && !node.Complete(s.focusDuration, s.breakDuration) {
		s.v.b.WriteString("\033[91;1m")
	}
	s.v.VisitNode(node)
}

func (s *ProgressPrinter) LeaveNode(node *PomConNode) {
	s.v.LeaveNode(node)
	if !s.currentPrinted && !node.Complete(s.focusDuration, s.breakDuration) {
		s.v.b.WriteString("\033[0m")
		s.currentPrinted = true
	}
}

func (s *ProgressPrinter) VisitComposite(node *PomConCompositeNode) {
	s.v.VisitComposite(node)
}

func (s *ProgressPrinter) VisitCompositeChild(node *PomConCompositeNode, i int) {
	s.v.VisitCompositeChild(node, i)
}

func (s *ProgressPrinter) LeaveCompositeChild(node *PomConCompositeNode, i int) {
	s.v.LeaveCompositeChild(node, i)
}

func (s *ProgressPrinter) LeaveComposite(node *PomConCompositeNode) {
	s.v.LeaveComposite(node)
}

func (s *ProgressPrinter) VisitRepeat(node *PomConRepeatNode) {
	if !s.currentPrinted && !node.Complete(s.focusDuration, s.breakDuration) {
		s.v.b.WriteString(fmt.Sprintf("%d/", node.repeats+1))
	}
	s.v.VisitRepeat(node)
}

func (s *ProgressPrinter) LeaveRepeat(node *PomConRepeatNode) {
	s.v.LeaveRepeat(node)
}

func NewProgressPrinter(focusDuration time.Duration, breakDuration time.Duration) *ProgressPrinter {
	return &ProgressPrinter{
		focusDuration:  focusDuration,
		breakDuration:  breakDuration,
		currentPrinted: false,
		v:              &PomConSimplePrinter{b: &strings.Builder{}},
	}
}
