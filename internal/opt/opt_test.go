package opt_test

import (
	"errors"
	"slices"
	"testing"

	"go-coreutils/internal/opt"
)

var (
	flagsABC = []opt.Option{
		{Short: 'a'},
		{Short: 'b'},
		{Short: 'c'},
	}

	headOptions = []opt.Option{
		{Short: 'n', TakesValue: true},
		{Short: 'c', TakesValue: true},
		{Short: 'q'},
		{Short: 'v'},
	}
)

// TestParse covers the shared scanner's success cases and its two usage
// errors. The error messages are prefixed with the program name.
func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		options      []opt.Option
		args         []string
		wantOperands []string
		wantOptions  []opt.Occurrence
		wantErr      string
	}{
		{name: "empty arguments"},
		{name: "operands only", args: []string{"fileA", "fileB"}, wantOperands: []string{"fileA", "fileB"}},
		{name: "single flag", options: flagsABC, args: []string{"-a"}, wantOptions: []opt.Occurrence{{Short: 'a'}}},
		{name: "flag cluster", options: flagsABC, args: []string{"-abc"}, wantOptions: []opt.Occurrence{{Short: 'a'}, {Short: 'b'}, {Short: 'c'}}},
		{name: "attached value", options: headOptions, args: []string{"-n5"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "5"}}},
		{name: "separate value", options: headOptions, args: []string{"-n", "5"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "5"}}},
		{name: "value may start with dash", options: headOptions, args: []string{"-n", "-5"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "-5"}}},
		{name: "value option at end of cluster", options: headOptions, args: []string{"-qn5"}, wantOptions: []opt.Occurrence{{Short: 'q'}, {Short: 'n', Value: "5"}}},
		{name: "value consumes rest of cluster", options: headOptions, args: []string{"-nq"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "q"}}},
		{name: "option operand permutation", options: headOptions, args: []string{"a.txt", "-n", "1", "b.txt"}, wantOperands: []string{"a.txt", "b.txt"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "1"}}},
		{name: "double dash stops options", options: flagsABC, args: []string{"-a", "--", "-b", "file"}, wantOperands: []string{"-b", "file"}, wantOptions: []opt.Occurrence{{Short: 'a'}}},
		{name: "single dash is operand", options: flagsABC, args: []string{"-", "-a"}, wantOperands: []string{"-"}, wantOptions: []opt.Occurrence{{Short: 'a'}}},
		{name: "preserves option order", options: headOptions, args: []string{"-n", "1", "-c", "3"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "1"}, {Short: 'c', Value: "3"}}},
		{name: "empty separate value", options: headOptions, args: []string{"-n", ""}, wantOptions: []opt.Occurrence{{Short: 'n', Value: ""}}},
		{name: "value consumes double dash", options: headOptions, args: []string{"-n", "--"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "--"}}},
		{name: "double dash alone", options: flagsABC, args: []string{"--"}},
		{name: "repeated flags preserved", options: flagsABC, args: []string{"-a", "-a"}, wantOptions: []opt.Occurrence{{Short: 'a'}, {Short: 'a'}}},
		{name: "operands after value", options: headOptions, args: []string{"-n", "2", "file"}, wantOperands: []string{"file"}, wantOptions: []opt.Occurrence{{Short: 'n', Value: "2"}}},
		{name: "cluster of flags then operand", options: flagsABC, args: []string{"-ab", "file"}, wantOperands: []string{"file"}, wantOptions: []opt.Occurrence{{Short: 'a'}, {Short: 'b'}}},

		{name: "unknown option", options: flagsABC, args: []string{"-z"}, wantErr: "head: unrecognized option: -z"},
		{name: "unknown character in cluster", options: flagsABC, args: []string{"-az"}, wantErr: "head: unrecognized option: -az"},
		{name: "missing value at end", options: headOptions, args: []string{"-n"}, wantErr: "head: option requires an argument: -n"},
		{name: "missing value after cluster flag", options: headOptions, args: []string{"-qn"}, wantErr: "head: option requires an argument: -n"},
		{name: "long option is unknown", options: flagsABC, args: []string{"--lines=5"}, wantErr: "head: unrecognized option: --lines=5"},
		{name: "unknown option after operand", options: flagsABC, args: []string{"file", "-z"}, wantErr: "head: unrecognized option: -z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := opt.Parse("head", tt.options, tt.args)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Parse() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("Parse() error = %q, want %q", err, tt.wantErr)
				}
				if _, ok := errors.AsType[*opt.UsageError](err); !ok {
					t.Fatalf("Parse() error type = %T, want *opt.UsageError", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}
			if !slices.Equal(got.Operands, tt.wantOperands) {
				t.Fatalf("Operands = %#v, want %#v", got.Operands, tt.wantOperands)
			}
			if !slices.Equal(got.Options, tt.wantOptions) {
				t.Fatalf("Options = %#v, want %#v", got.Options, tt.wantOptions)
			}
		})
	}
}
