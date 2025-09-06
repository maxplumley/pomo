package main

import (
	"testing"
	"time"
)

func shouldParse(str string, t *testing.T) {
	node, err := FromString(str)
	if err != nil {
		t.Errorf("unexpected error parsing valid PomCon %s: %v", str, err)
	}
	if node == nil {
		t.Errorf("unexpected nil result from parsing PomCon %s", str)
	}
}

func shouldNotParse(str string, t *testing.T) {
	_, err := FromString(str)
	if err == nil {
		t.Errorf("expected error parsing invalid PomCon %s", str)
	}
}

func TestParsePomCon(t *testing.T) {
	tests := []string{
		"f",
		"b",
		"p",
		"2f",
		"2b",
		"2h2m2sf",
		"2fp",
		"2b2fp",
		"2:f",
		"2:2h2m2sf",
		"2:p",
		"2:2f2h2m2sbp",
		"b f",
		"b  f",
		"2:2b f",
		"2[2f]",
		"2[2f ]",
		"2[ 2f ]",
		"2[2f 2b]",
		"[2f 2b] 2f",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			shouldParse(tt, t)
		})
	}
}

func TestParseInvalidPomCon(t *testing.T) {
	tests := []string{
		"",
		"2",
		"2:",
		"2h",
		"2h2m2s",
		"2:2h",
		"2: 2f",
		"2 [2hf]",
		"2:[2f]",
		"2[2f 2b",
		"2f 2b]",
		"0f",
		"0:f",
		"0:0h0m0sf",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			shouldNotParse(tt, t)
		})
	}
}

func TestPrintPomCon(t *testing.T) {
	tests := []struct {
		p PomCon
		s string
	}{
		{
			p: &PomConNode{
				focusDuration: DURATION_DEFAULT,
				breakDuration: DURATION_NOT_SET,
			},
			s: "f",
		},
		{
			p: &PomConNode{
				focusDuration: DURATION_NOT_SET,
				breakDuration: DURATION_DEFAULT,
			},
			s: "b",
		},
		{
			p: &PomConNode{
				focusDuration: DURATION_DEFAULT,
				breakDuration: DURATION_DEFAULT,
			},
			s: "p",
		},
		{
			p: &PomConNode{
				focusDuration: time.Duration(2) * time.Minute,
				breakDuration: DURATION_NOT_SET,
			},
			s: "2f",
		},
		{
			p: &PomConNode{
				focusDuration: DURATION_NOT_SET,
				breakDuration: time.Duration(2) * time.Minute,
			},
			s: "2b",
		},
		{
			p: &PomConNode{
				focusDuration: time.Duration(2)*time.Hour + time.Duration(2)*time.Minute + time.Duration(2)*time.Second,
				breakDuration: DURATION_NOT_SET,
			},
			s: "2h2m2sf",
		},
		{
			p: &PomConNode{
				focusDuration: DURATION_NOT_SET,
				breakDuration: time.Duration(2)*time.Hour + time.Duration(2)*time.Minute + time.Duration(2)*time.Second,
			},
			s: "2h2m2sb",
		},
		{
			p: &PomConNode{
				focusDuration: time.Duration(2) * time.Hour,
				breakDuration: DURATION_NOT_SET,
			},
			s: "2hf",
		},
		{
			p: &PomConNode{
				focusDuration: DURATION_DEFAULT,
				breakDuration: time.Duration(2) * time.Hour,
			},
			s: "2hbp",
		},
		{
			p: &PomConNode{
				focusDuration: time.Duration(2) * time.Hour,
				breakDuration: DURATION_DEFAULT,
			},
			s: "2hfp",
		},
		{
			p: &PomConNode{
				focusDuration: time.Duration(2)*time.Hour + time.Duration(2)*time.Minute + time.Duration(2)*time.Second,
				breakDuration: time.Duration(2)*time.Hour + time.Duration(2)*time.Minute + time.Duration(2)*time.Second,
			},
			s: "2h2m2sf2h2m2sbp",
		},
		{
			p: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: time.Duration(2) * time.Hour,
						breakDuration: DURATION_NOT_SET,
					},
				},
			},
			s: "2hf",
		},
		{
			p: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: time.Duration(2) * time.Hour,
						breakDuration: DURATION_NOT_SET,
					},
					&PomConNode{
						focusDuration: time.Duration(2) * time.Hour,
						breakDuration: DURATION_NOT_SET,
					},
				},
			},
			s: "2hf 2hf",
		},
		{
			p: &PomConRepeatNode{
				repeat: 2,
				child: &PomConNode{
					focusDuration: time.Duration(2) * time.Hour,
					breakDuration: DURATION_NOT_SET,
				},
			},
			s: "2:2hf",
		},
		{
			p: &PomConRepeatNode{
				repeat: 2,
				child: &PomConCompositeNode{
					children: []PomCon{
						&PomConNode{
							focusDuration: time.Duration(2) * time.Hour,
							breakDuration: DURATION_NOT_SET,
						},
					},
				},
			},
			s: "2[2hf]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			var str, err = Print(tt.p)
			if err != nil {
				t.Errorf("unexpected error printing PomCon: %v", err)
				return
			}
			if tt.s != str {
				t.Errorf("expected %s to equal %s", str, tt.s)
			}
		})
	}
}

func TestTickPomConNode(t *testing.T) {
	tests := []struct {
		t  string
		n  *PomConNode
		df time.Duration
		db time.Duration
		d  time.Duration
		r  time.Duration
	}{
		{
			t: "no remainder",
			n: &PomConNode{
				focusDuration: 25 * time.Minute,
				breakDuration: 5 * time.Minute,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  29 * time.Minute,
			r:  0,
		},
		{
			t: "no remainder boundary",
			n: &PomConNode{
				focusDuration: 25 * time.Minute,
				breakDuration: 5 * time.Minute,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  30 * time.Minute,
			r:  0,
		},
		{
			t: "1 minute remainder boundary",
			n: &PomConNode{
				focusDuration: 25 * time.Minute,
				breakDuration: 5 * time.Minute,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  31 * time.Minute,
			r:  1 * time.Minute,
		},
		{
			t: "default focus duration, no remainder",
			n: &PomConNode{
				focusDuration: DURATION_DEFAULT,
				breakDuration: 5 * time.Minute,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  34 * time.Minute,
			r:  0,
		},
		{
			t: "default focus duration, no remainder boundary",
			n: &PomConNode{
				focusDuration: DURATION_DEFAULT,
				breakDuration: 5 * time.Minute,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  35 * time.Minute,
			r:  0,
		},
		{
			t: "default focus duration, remainder boundary",
			n: &PomConNode{
				focusDuration: DURATION_DEFAULT,
				breakDuration: 5 * time.Minute,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  36 * time.Minute,
			r:  1 * time.Minute,
		},
		{
			t: "default break duration, no remainder",
			n: &PomConNode{
				focusDuration: 25 * time.Minute,
				breakDuration: DURATION_DEFAULT,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  34 * time.Minute,
			r:  0,
		},
		{
			t: "default break duration, no remainder boundary",
			n: &PomConNode{
				focusDuration: 25 * time.Minute,
				breakDuration: DURATION_DEFAULT,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  35 * time.Minute,
			r:  0,
		},
		{
			t: "default break duration, 1 minute remainder boundary",
			n: &PomConNode{
				focusDuration: 25 * time.Minute,
				breakDuration: DURATION_DEFAULT,
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  36 * time.Minute,
			r:  1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.t, func(t *testing.T) {
			rem := tt.n.Tick(tt.d, tt.df, tt.db)
			if rem != tt.r {
				t.Errorf("expected remainder %v, but got %v", tt.r, rem)
			}
		})
	}
}

func TestTickPomConCompositeNode(t *testing.T) {
	tests := []struct {
		t  string
		n  *PomConCompositeNode
		df time.Duration
		db time.Duration
		d  time.Duration
		r  time.Duration
	}{
		{
			t: "single child, no remainder",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  29 * time.Minute,
			r:  0,
		},
		{
			t: "single child, no remainder boundary",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  30 * time.Minute,
			r:  0,
		},
		{
			t: "single child, 1 minute remainder boundary",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  31 * time.Minute,
			r:  1 * time.Minute,
		},
		{
			t: "single child, defaults",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: DURATION_DEFAULT,
						breakDuration: DURATION_DEFAULT,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  41 * time.Minute,
			r:  1 * time.Minute,
		},
		{
			t: "two children, no remainder",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  59 * time.Minute,
			r:  0,
		},
		{
			t: "two children, no remainder boundary",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  60 * time.Minute,
			r:  0,
		},
		{
			t: "two children, 1 minute remainder boundary",
			n: &PomConCompositeNode{
				children: []PomCon{
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
					&PomConNode{
						focusDuration: 25 * time.Minute,
						breakDuration: 5 * time.Minute,
					},
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  61 * time.Minute,
			r:  1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.t, func(t *testing.T) {
			rem := tt.n.Tick(tt.d, tt.df, tt.db)
			if rem != tt.r {
				t.Errorf("expected remainder %v, but got %v", tt.r, rem)
			}
		})
	}
}

func TestTickPomConRepeatNode(t *testing.T) {
	tests := []struct {
		t  string
		n  *PomConRepeatNode
		df time.Duration
		db time.Duration
		d  time.Duration
		r  time.Duration
	}{
		{
			t: "single repeat, no remainder",
			n: &PomConRepeatNode{
				repeat: 1,
				child: &PomConNode{
					focusDuration: 25 * time.Minute,
					breakDuration: 5 * time.Minute,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  29 * time.Minute,
			r:  0,
		},
		{
			t: "single repeat, no remainder boundary",
			n: &PomConRepeatNode{
				repeat: 1,
				child: &PomConNode{
					focusDuration: 25 * time.Minute,
					breakDuration: 5 * time.Minute,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  30 * time.Minute,
			r:  0,
		},
		{
			t: "single repeat, 1 minute remainder boundary",
			n: &PomConRepeatNode{
				repeat: 1,
				child: &PomConNode{
					focusDuration: 25 * time.Minute,
					breakDuration: 5 * time.Minute,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  31 * time.Minute,
			r:  1 * time.Minute,
		},
		{
			t: "single repeat, defaults",
			n: &PomConRepeatNode{
				repeat: 1,
				child: &PomConNode{
					focusDuration: DURATION_DEFAULT,
					breakDuration: DURATION_DEFAULT,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  41 * time.Minute,
			r:  1 * time.Minute,
		},
		{
			t: "two repeats, no remainder",
			n: &PomConRepeatNode{
				repeat: 2,
				child: &PomConNode{
					focusDuration: 25 * time.Minute,
					breakDuration: 5 * time.Minute,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  59 * time.Minute,
			r:  0,
		},
		{
			t: "two repeats, no remainder boundary",
			n: &PomConRepeatNode{
				repeat: 2,
				child: &PomConNode{
					focusDuration: 25 * time.Minute,
					breakDuration: 5 * time.Minute,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  60 * time.Minute,
			r:  0,
		},
		{
			t: "two repeats, 1 minute remainder boundary",
			n: &PomConRepeatNode{
				repeat: 2,
				child: &PomConNode{
					focusDuration: 25 * time.Minute,
					breakDuration: 5 * time.Minute,
				},
			},
			df: 30 * time.Minute,
			db: 10 * time.Minute,
			d:  61 * time.Minute,
			r:  1 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.t, func(t *testing.T) {
			rem := tt.n.Tick(tt.d, tt.df, tt.db)
			if rem != tt.r {
				t.Errorf("expected remainder %v, but got %v", tt.r, rem)
			}
		})
	}
}
