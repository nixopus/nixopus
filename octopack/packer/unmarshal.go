package packer

import (
	"encoding/json"
	"fmt"
)

// unmarshalInstruction unmarshals a single instruction from JSON.
// This is used internally by Stage.UnmarshalJSON to avoid circular imports.
func unmarshalInstruction(data []byte) (Instruction, error) {
	var wrapper instructionWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	switch wrapper.Type {
	case "run":
		var r RunInstruction
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, err
		}
		return r, nil
	case "cmd":
		var c CmdInstruction
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, err
		}
		return c, nil
	case "entrypoint":
		var e EntrypointInstruction
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "copy":
		var c CopyInstruction
		if err := json.Unmarshal(data, &c); err != nil {
			return nil, err
		}
		return c, nil
	case "add":
		var a AddInstruction
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "env":
		var e EnvInstruction
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "arg":
		var a ArgInstruction
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return a, nil
	case "expose":
		var e ExposeInstruction
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}
		return e, nil
	case "volume":
		var v VolumeInstruction
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return v, nil
	case "user":
		var u UserInstruction
		if err := json.Unmarshal(data, &u); err != nil {
			return nil, err
		}
		return u, nil
	case "workdir":
		var w WorkdirInstruction
		if err := json.Unmarshal(data, &w); err != nil {
			return nil, err
		}
		return w, nil
	case "label":
		var l LabelInstruction
		if err := json.Unmarshal(data, &l); err != nil {
			return nil, err
		}
		return l, nil
	case "stopsignal":
		var s StopsignalInstruction
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		return s, nil
	case "healthcheck":
		var h HealthcheckInstruction
		if err := json.Unmarshal(data, &h); err != nil {
			return nil, err
		}
		return h, nil
	case "shell":
		var s ShellInstruction
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		return s, nil
	case "onbuild":
		var o OnbuildInstruction
		var onbuildData struct {
			Instruction json.RawMessage `json:"instruction"`
		}
		if err := json.Unmarshal(data, &onbuildData); err != nil {
			return nil, err
		}
		nestedInst, err := unmarshalInstruction(onbuildData.Instruction)
		if err != nil {
			return nil, err
		}
		o.Instruction = nestedInst
		return o, nil
	case "maintainer":
		var m MaintainerInstruction
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unknown instruction type: %s", wrapper.Type)
	}
}

// UnmarshalJSON implements custom unmarshaling for Stage to handle Instruction discriminated union.
func (s *Stage) UnmarshalJSON(data []byte) error {
	var stage struct {
		From         FromInstruction `json:"from"`
		Instructions json.RawMessage `json:"instructions,omitempty"`
	}

	if err := json.Unmarshal(data, &stage); err != nil {
		return err
	}

	s.From = stage.From

	if len(stage.Instructions) == 0 {
		s.Instructions = nil
		return nil
	}

	// Parse instructions array
	var rawInstructions []json.RawMessage
	if err := json.Unmarshal(stage.Instructions, &rawInstructions); err != nil {
		return err
	}

	// Unmarshal each instruction
	s.Instructions = make([]Instruction, 0, len(rawInstructions))
	for _, raw := range rawInstructions {
		inst, err := unmarshalInstruction(raw)
		if err != nil {
			return err
		}
		s.Instructions = append(s.Instructions, inst)
	}

	return nil
}
