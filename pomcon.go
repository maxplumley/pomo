package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type PomConNodeType = int
type PomConSessionType = int

const (
	FocusSession PomConSessionType = iota
	BreakSession
)

type PomConSession struct {
	t        PomConSessionType
	duration time.Duration
	elapsed  time.Duration
}

const (
	DURATION_DEFAULT = 0
	DURATION_NOT_SET = -1
	REPEAT_DEFAULT   = 1
	REPEAT_NOT_SET   = -1
)

type PomConNode struct {
	focusDuration time.Duration
	breakDuration time.Duration
	duration      time.Duration
}

func durationOrDefault(duration time.Duration, defaultDuration time.Duration) time.Duration {
	if duration == DURATION_NOT_SET {
		return 0
	} else if duration == DURATION_DEFAULT {
		return defaultDuration
	}
	return duration
}

func (s *PomConNode) Tick(duration time.Duration, focusDuration time.Duration, breakDuration time.Duration) time.Duration {
	fd := durationOrDefault(s.focusDuration, focusDuration)
	bd := durationOrDefault(s.breakDuration, breakDuration)

	durationTotal := fd + bd
	durationLeft := durationTotal - s.duration
	if durationLeft >= duration {
		s.duration += duration
		return 0
	} else {
		s.duration = durationTotal
		return duration - durationLeft
	}
}

func (s *PomConNode) Current(focusDuration time.Duration, breakDuration time.Duration) PomConSession {
	fd := durationOrDefault(s.focusDuration, focusDuration)
	if s.duration < fd {
		return PomConSession{
			t:        FocusSession,
			duration: fd,
			elapsed:  s.duration,
		}
	}
	bd := durationOrDefault(s.breakDuration, breakDuration)
	return PomConSession{
		t:        BreakSession,
		duration: bd,
		elapsed:  s.duration - fd,
	}
}

func (s *PomConNode) Reset() {
	s.duration = 0
}

func (s *PomConNode) Complete(focusDuration time.Duration, breakDuration time.Duration) bool {
	return s.duration == s.Duration(focusDuration, breakDuration)
}

func (s *PomConNode) Duration(focusDuration time.Duration, breakDuration time.Duration) time.Duration {
	return durationOrDefault(s.focusDuration, focusDuration) + durationOrDefault(s.breakDuration, breakDuration)
}

type PomConCompositeNode struct {
	parent   *PomConCompositeNode
	children []PomCon
}

func (s *PomConCompositeNode) Tick(duration time.Duration, focusDuration time.Duration, breakDuration time.Duration) time.Duration {
	remainder := duration
	for _, c := range s.children {
		remainder = c.Tick(remainder, focusDuration, breakDuration)
		if remainder == 0 {
			break
		}
	}
	return remainder
}

func (s *PomConCompositeNode) Current(focusDuration time.Duration, breakDuration time.Duration) PomConSession {
	for i, c := range s.children {
		if !c.Complete(focusDuration, breakDuration) || i == len(s.children)-1 {
			return c.Current(focusDuration, breakDuration)
		}
	}

	// should never be called, len(s.children) must be > 0
	return PomConSession{
		t:        FocusSession,
		duration: 0,
		elapsed:  0,
	}
}

func (s *PomConCompositeNode) Reset() {
	for _, c := range s.children {
		c.Reset()
	}
}

func (s *PomConCompositeNode) Complete(focusDuration time.Duration, breakDuration time.Duration) bool {
	complete := true
	for _, c := range s.children {
		complete = complete && c.Complete(focusDuration, breakDuration)
	}
	return complete
}

func (s *PomConCompositeNode) Duration(focusDuration time.Duration, breakDuration time.Duration) time.Duration {
	total := 0 * time.Second
	for _, c := range s.children {
		total += c.Duration(focusDuration, breakDuration)
	}
	return total
}

type PomConRepeatNode struct {
	repeat  int
	repeats int
	child   PomCon
}

func (s *PomConRepeatNode) Tick(duration time.Duration, focusDuration time.Duration, breakDuration time.Duration) time.Duration {
	remainder := duration
	for !s.Complete(focusDuration, breakDuration) && remainder > 0 {
		remainder = s.child.Tick(remainder, focusDuration, breakDuration)
		if remainder > 0 && s.repeats < s.repeat-1 {
			s.repeats++
			s.child.Reset()
		}
	}
	return remainder
}

func (s *PomConRepeatNode) Current(focusDuration time.Duration, breakDuration time.Duration) PomConSession {
	return s.child.Current(focusDuration, breakDuration)
}

func (s *PomConRepeatNode) Reset() {
	s.repeats = 0
	s.child.Reset()
}

func (s *PomConRepeatNode) Complete(focusDuration time.Duration, breakDuration time.Duration) bool {
	return s.repeats == s.repeat-1 && s.child.Complete(focusDuration, breakDuration)
}

func (s *PomConRepeatNode) Duration(focusDuration time.Duration, breakDuration time.Duration) time.Duration {
	return time.Duration(s.repeat) * s.child.Duration(focusDuration, breakDuration)
}

type PomCon interface {
	Tick(duration time.Duration, focusDuration time.Duration, breakDuration time.Duration) time.Duration
	Current(focusDuration time.Duration, breakDuration time.Duration) PomConSession
	Reset()
	Complete(focusDuration time.Duration, breakDuration time.Duration) bool
	Duration(focusDuration time.Duration, breakDuration time.Duration) time.Duration
}

type PomConVisitor interface {
	VisitNode(node *PomConNode)
	LeaveNode(node *PomConNode)
	VisitComposite(node *PomConCompositeNode)
	VisitCompositeChild(node *PomConCompositeNode, i int)
	LeaveCompositeChild(node *PomConCompositeNode, i int)
	LeaveComposite(node *PomConCompositeNode)
	VisitRepeat(node *PomConRepeatNode)
	LeaveRepeat(node *PomConRepeatNode)
}

func Crawl(pomcon PomCon, visitor PomConVisitor) error {
	switch node := pomcon.(type) {
	case *PomConNode:
		visitor.VisitNode(node)
		visitor.LeaveNode(node)
	case *PomConRepeatNode:
		visitor.VisitRepeat(node)
		err := Crawl(node.child, visitor)
		if err != nil {
			return err
		}
		visitor.LeaveRepeat(node)
	case *PomConCompositeNode:
		visitor.VisitComposite(node)
		for i, child := range node.children {
			visitor.VisitCompositeChild(node, i)
			err := Crawl(child, visitor)
			if err != nil {
				return err
			}
			visitor.LeaveCompositeChild(node, i)
		}
		visitor.LeaveComposite(node)
	default:
		return fmt.Errorf("unexpected PomCon implementation %s", reflect.TypeOf(node))
	}
	return nil
}

type PomConSimplePrinter struct {
	b *strings.Builder
}

func (s *PomConSimplePrinter) VisitNode(node *PomConNode) {
	writeDuration := func(d time.Duration) {
		if d <= DURATION_DEFAULT {
			return
		}

		hours := d / time.Hour
		minutes := d % time.Hour / time.Minute
		seconds := d % time.Minute / time.Second

		if hours+seconds == 0 {
			s.b.WriteString(fmt.Sprintf("%d", minutes))
			return
		}

		if hours > 0 {
			s.b.WriteString(fmt.Sprintf("%dh", hours))
		}
		if minutes > 0 {
			s.b.WriteString(fmt.Sprintf("%dm", minutes))
		}
		if seconds > 0 {
			s.b.WriteString(fmt.Sprintf("%ds", seconds))
		}
	}

	if node.focusDuration > DURATION_DEFAULT {
		writeDuration(node.focusDuration)
		s.b.WriteString("f")
	} else if node.focusDuration == DURATION_DEFAULT && node.breakDuration == DURATION_NOT_SET {
		s.b.WriteString("f")
	}

	if node.breakDuration > DURATION_DEFAULT {
		writeDuration(node.breakDuration)
		s.b.WriteString("b")
	} else if node.breakDuration == DURATION_DEFAULT && node.focusDuration == DURATION_NOT_SET {
		s.b.WriteString("b")
	}

	if node.focusDuration != DURATION_NOT_SET && node.breakDuration != DURATION_NOT_SET {
		s.b.WriteString("p")
	}
}

func (s *PomConSimplePrinter) LeaveNode(node *PomConNode) {}

func (s *PomConSimplePrinter) VisitComposite(node *PomConCompositeNode) {}

func (s *PomConSimplePrinter) VisitCompositeChild(node *PomConCompositeNode, i int) {
	if i > 0 {
		s.b.WriteString(" ")
	}
}

func (s *PomConSimplePrinter) LeaveCompositeChild(node *PomConCompositeNode, i int) {}

func (s *PomConSimplePrinter) LeaveComposite(node *PomConCompositeNode) {}

func (s *PomConSimplePrinter) VisitRepeat(node *PomConRepeatNode) {
	switch node.child.(type) {
	case *PomConNode:
		s.b.WriteString(fmt.Sprintf("%d:", node.repeat))
	case *PomConRepeatNode:
		s.b.WriteString(fmt.Sprintf("%d:", node.repeat))
	default:
		s.b.WriteString(fmt.Sprintf("%d[", node.repeat))
	}
}

func (s *PomConSimplePrinter) LeaveRepeat(node *PomConRepeatNode) {
	switch node.child.(type) {
	case *PomConNode:
	case *PomConRepeatNode:
	default:
		s.b.WriteString("]")
	}
}

func Print(pomcon PomCon) (string, error) {
	p := &PomConSimplePrinter{b: &strings.Builder{}}
	err := Crawl(pomcon, p)
	if err != nil {
		return "", err
	}
	return p.b.String(), nil
}

func unexpectedToken(c rune, i int) error {
	return fmt.Errorf("unexpected '%c' at postition %d", c, i)
}

func numberError(number string, i int) error {
	return fmt.Errorf("internal error at position %d, expected %s to be an integer", i, number)
}

func FromString(pomcon string) (PomCon, error) {
	depth := 0
	number := ""
	repeat := REPEAT_NOT_SET
	var duration time.Duration = DURATION_NOT_SET
	var focusDuration time.Duration = DURATION_NOT_SET
	var breakDuration time.Duration = DURATION_NOT_SET
	root := &PomConCompositeNode{}
	current := root
	for i, c := range pomcon {
		if unicode.IsDigit(c) {
			number = number + string(c)
		} else if c == '[' {
			depth++
			if repeat != REPEAT_NOT_SET {
				return nil, unexpectedToken(c, i)
			}
			if number != "" {
				var err error
				repeat, err = strconv.Atoi(number)
				if err != nil {
					return nil, numberError(number, i)
				}
				number = ""
			}
			node := &PomConCompositeNode{
				parent:   current,
				children: make([]PomCon, 0),
			}
			var child PomCon = node
			if repeat == 0 {
				return nil, fmt.Errorf("invalid zero repeat at position %d", i)
			}
			if repeat > REPEAT_DEFAULT {
				child = &PomConRepeatNode{
					repeat: repeat,
					child:  node,
				}
			}
			current.children = append(current.children, child)
			current = node
			repeat = REPEAT_NOT_SET
		} else if c == ':' {
			if number == "" {
				return nil, unexpectedToken(c, i)
			}
			var err error
			repeat, err = strconv.Atoi(number)
			if err != nil {
				return nil, numberError(number, i)
			}
			number = ""
		} else if c == 'h' {
			if number == "" {
				return nil, unexpectedToken(c, i)
			}
			hours, err := strconv.Atoi(number)
			if err != nil {
				return nil, numberError(number, i)
			}
			duration = max(0, duration) + time.Duration(hours)*time.Hour
			number = ""
		} else if c == 'm' {
			if number == "" {
				return nil, unexpectedToken(c, i)
			}
			minutes, err := strconv.Atoi(number)
			if err != nil {
				return nil, numberError(number, i)
			}
			duration = max(0, duration) + time.Duration(minutes)*time.Minute
			number = ""
		} else if c == 's' {
			if number == "" {
				return nil, unexpectedToken(c, i)
			}
			seconds, err := strconv.Atoi(number)
			if err != nil {
				return nil, numberError(number, i)
			}
			duration = max(0, duration) + time.Duration(seconds)*time.Second
			number = ""
		} else if c == 'b' {
			if breakDuration != DURATION_NOT_SET {
				return nil, unexpectedToken(c, i)
			}
			if duration == 0 {
				return nil, fmt.Errorf("invalid zero break duration at position %d", i)
			}
			if duration > 0 && number != "" {
				return nil, unexpectedToken(c, i)
			} else if number != "" {
				minutes, err := strconv.Atoi(number)
				if err != nil {
					return nil, numberError(number, i)
				}
				if minutes <= 0 {
					return nil, fmt.Errorf("invalid zero break duration at position %d", i)
				}
				breakDuration = time.Duration(minutes) * time.Minute
			} else {
				breakDuration = max(DURATION_DEFAULT, duration)
			}
			duration = DURATION_NOT_SET
			number = ""
		} else if c == 'f' {
			if focusDuration != DURATION_NOT_SET {
				return nil, unexpectedToken(c, i)
			}
			if duration == 0 {
				return nil, fmt.Errorf("invalid zero focus duration at position %d", i)
			}
			if duration > 0 && number != "" {
				return nil, unexpectedToken(c, i)
			} else if number != "" {
				minutes, err := strconv.Atoi(number)
				if err != nil {
					return nil, numberError(number, i)
				}
				if minutes <= 0 {
					return nil, fmt.Errorf("invalid zero focus duration at position %d", i)
				}
				focusDuration = time.Duration(minutes) * time.Minute
			} else {
				focusDuration = max(DURATION_DEFAULT, duration)
			}
			duration = DURATION_NOT_SET
			number = ""
		} else if c == 'p' {
			if duration != DURATION_NOT_SET || number != "" {
				return nil, unexpectedToken(c, i)
			}
			breakDuration = max(DURATION_DEFAULT, breakDuration)
			focusDuration = max(DURATION_DEFAULT, focusDuration)
		} else if c == ' ' || c == ']' {
		} else {
			return nil, unexpectedToken(c, i)
		}
		if c == ' ' || c == ']' || i == len(pomcon)-1 {
			if c == ']' {
				if depth > 0 {
					depth--
				} else {
					return nil, unexpectedToken(c, i)
				}
			}
			if repeat == 0 {
				return nil, fmt.Errorf("invalid zero repeat at position %d", i)
			}
			if number != "" || duration > 0 {
				if c == ' ' || c == ']' {
					return nil, unexpectedToken(c, i)
				} else {
					return nil, errors.New("unexpected end of expression")
				}
			}

			if c == ' ' && breakDuration == DURATION_NOT_SET && focusDuration == DURATION_NOT_SET && repeat != REPEAT_NOT_SET {
				return nil, unexpectedToken(c, i)
			}

			if c != ' ' && breakDuration == DURATION_NOT_SET && focusDuration == DURATION_NOT_SET && len(current.children) == 0 {
				return nil, unexpectedToken(c, i)
			}

			var node PomCon
			node = &PomConNode{
				focusDuration: focusDuration,
				breakDuration: breakDuration,
				duration:      0 * time.Second,
			}
			if repeat > REPEAT_DEFAULT {
				node = &PomConRepeatNode{
					repeat: repeat,
					child:  node,
				}
			}

			current.children = append(current.children, node)
			if c == ']' {
				current = current.parent
			}
			repeat = REPEAT_NOT_SET
			focusDuration = DURATION_NOT_SET
			breakDuration = DURATION_NOT_SET
		}
	}

	if depth > 0 {
		return nil, errors.New("unbalanced '[' detected")
	}

	if len(root.children) == 0 {
		return nil, errors.New("invalid empty PomCon expression")
	}

	return root, nil
}
