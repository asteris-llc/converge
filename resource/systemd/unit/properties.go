// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unit

// Properties defines the properties common to all unit structures
type Properties struct {
	RefuseManualStop                bool
	RebootArgument                  string
	ActiveState                     string
	ActiveEnterTimestamp            uint64
	Requires                        []string
	LoadError                       []interface{}
	CanReload                       bool
	StartLimitAction                string
	Transient                       bool
	OnFailure                       []string
	Asserts                         [][]interface{}
	RefuseManualStart               bool
	WantedBy                        []string
	LoadState                       string
	BoundBy                         []string
	DropInPaths                     []string
	Conflicts                       []string
	Documentation                   []string
	Wants                           []string
	IgnoreOnIsolate                 bool
	ActiveEnterTimestampMonotonic   uint64
	InactiveEnterTimestampMonotonic uint64
	Requisite                       []string
	BindsTo                         []string
	Job                             []interface{}
	ConflictedBy                    []string
	Before                          []string
	AssertTimestampMonotonic        uint64
	ActiveExitTimestampMonotonic    uint64
	ConsistsOf                      []string
	CanStart                        bool
	CanStop                         bool
	RequiresMountsFor               []string
	TriggeredBy                     []string
	StartLimitInterval              uint64
	JobTimeoutAction                string
	ConditionTimestamp              uint64
	InactiveExitTimestamp           uint64
	OnFailureJobMode                string
	CanIsolate                      bool
	InactiveEnterTimestamp          uint64
	ConditionTimestampMonotonic     uint64
	StartLimitBurst                 uint32
	PartOf                          []string
	Conditions                      [][]interface{}
	DefaultDependencies             bool
	RequiredBy                      []string
	After                           []string
	FragmentPath                    string
	UnitFilePreset                  string
	ConditionResult                 bool
	Description                     string
	Triggers                        []string
	InactiveExitTimestampMonotonic  uint64
	ReloadPropagatedFrom            []string
	SourcePath                      string
	Id                              string
	JobTimeoutUSec                  uint64
	AssertResult                    bool
	ActiveExitTimestamp             uint64
	RequisiteOf                     []string
	StateChangeTimestampMonotonic   uint64
	StateChangeTimestamp            uint64
	PropagatesReloadTo              []string
	Names                           []string
	NeedDaemonReload                bool
	AssertTimestamp                 uint64
	AllowIsolate                    bool
	JobTimeoutRebootArgument        string
	UnitFileState                   string
	StopWhenUnneeded                bool
	JoinsNamespaceOf                []string
	SubState                        string
	Following                       string
}

func newFromMap(m map[string]interface{}) *Properties {
	p := &Properties{}
	if val, ok := m["RefuseManualStop"]; ok {
		p.RefuseManualStop = val.(bool)
	}
	if val, ok := m["RebootArgument"]; ok {
		p.RebootArgument = val.(string)
	}
	if val, ok := m["ActiveState"]; ok {
		p.ActiveState = val.(string)
	}
	if val, ok := m["ActiveEnterTimestamp"]; ok {
		p.ActiveEnterTimestamp = val.(uint64)
	}
	if val, ok := m["Requires"]; ok {
		p.Requires = val.([]string)
	}
	if val, ok := m["LoadError"]; ok {
		p.LoadError = val.([]interface{})
	}
	if val, ok := m["CanReload"]; ok {
		p.CanReload = val.(bool)
	}
	if val, ok := m["StartLimitAction"]; ok {
		p.StartLimitAction = val.(string)
	}
	if val, ok := m["Transient"]; ok {
		p.Transient = val.(bool)
	}
	if val, ok := m["OnFailure"]; ok {
		p.OnFailure = val.([]string)
	}
	if val, ok := m["Asserts"]; ok {
		p.Asserts = val.([][]interface{})
	}
	if val, ok := m["RefuseManualStart"]; ok {
		p.RefuseManualStart = val.(bool)
	}
	if val, ok := m["WantedBy"]; ok {
		p.WantedBy = val.([]string)
	}
	if val, ok := m["LoadState"]; ok {
		p.LoadState = val.(string)
	}
	if val, ok := m["BoundBy"]; ok {
		p.BoundBy = val.([]string)
	}
	if val, ok := m["DropInPaths"]; ok {
		p.DropInPaths = val.([]string)
	}
	if val, ok := m["Conflicts"]; ok {
		p.Conflicts = val.([]string)
	}
	if val, ok := m["Documentation"]; ok {
		p.Documentation = val.([]string)
	}
	if val, ok := m["Wants"]; ok {
		p.Wants = val.([]string)
	}
	if val, ok := m["IgnoreOnIsolate"]; ok {
		p.IgnoreOnIsolate = val.(bool)
	}
	if val, ok := m["ActiveEnterTimestampMonotonic"]; ok {
		p.ActiveEnterTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["InactiveEnterTimestampMonotonic"]; ok {
		p.InactiveEnterTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["Requisite"]; ok {
		p.Requisite = val.([]string)
	}
	if val, ok := m["BindsTo"]; ok {
		p.BindsTo = val.([]string)
	}
	if val, ok := m["Job"]; ok {
		p.Job = val.([]interface{})
	}
	if val, ok := m["ConflictedBy"]; ok {
		p.ConflictedBy = val.([]string)
	}
	if val, ok := m["Before"]; ok {
		p.Before = val.([]string)
	}
	if val, ok := m["AssertTimestampMonotonic"]; ok {
		p.AssertTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["ActiveExitTimestampMonotonic"]; ok {
		p.ActiveExitTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["ConsistsOf"]; ok {
		p.ConsistsOf = val.([]string)
	}
	if val, ok := m["CanStart"]; ok {
		p.CanStart = val.(bool)
	}
	if val, ok := m["CanStop"]; ok {
		p.CanStop = val.(bool)
	}
	if val, ok := m["RequiresMountsFor"]; ok {
		p.RequiresMountsFor = val.([]string)
	}
	if val, ok := m["TriggeredBy"]; ok {
		p.TriggeredBy = val.([]string)
	}
	if val, ok := m["StartLimitInterval"]; ok {
		p.StartLimitInterval = val.(uint64)
	}
	if val, ok := m["JobTimeoutAction"]; ok {
		p.JobTimeoutAction = val.(string)
	}
	if val, ok := m["ConditionTimestamp"]; ok {
		p.ConditionTimestamp = val.(uint64)
	}
	if val, ok := m["InactiveExitTimestamp"]; ok {
		p.InactiveExitTimestamp = val.(uint64)
	}
	if val, ok := m["OnFailureJobMode"]; ok {
		p.OnFailureJobMode = val.(string)
	}
	if val, ok := m["CanIsolate"]; ok {
		p.CanIsolate = val.(bool)
	}
	if val, ok := m["InactiveEnterTimestamp"]; ok {
		p.InactiveEnterTimestamp = val.(uint64)
	}
	if val, ok := m["ConditionTimestampMonotonic"]; ok {
		p.ConditionTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["StartLimitBurst"]; ok {
		p.StartLimitBurst = val.(uint32)
	}
	if val, ok := m["PartOf"]; ok {
		p.PartOf = val.([]string)
	}
	if val, ok := m["Conditions"]; ok {
		p.Conditions = val.([][]interface{})
	}
	if val, ok := m["DefaultDependencies"]; ok {
		p.DefaultDependencies = val.(bool)
	}
	if val, ok := m["RequiredBy"]; ok {
		p.RequiredBy = val.([]string)
	}
	if val, ok := m["After"]; ok {
		p.After = val.([]string)
	}
	if val, ok := m["FragmentPath"]; ok {
		p.FragmentPath = val.(string)
	}
	if val, ok := m["UnitFilePreset"]; ok {
		p.UnitFilePreset = val.(string)
	}
	if val, ok := m["ConditionResult"]; ok {
		p.ConditionResult = val.(bool)
	}
	if val, ok := m["Description"]; ok {
		p.Description = val.(string)
	}
	if val, ok := m["Triggers"]; ok {
		p.Triggers = val.([]string)
	}
	if val, ok := m["InactiveExitTimestampMonotonic"]; ok {
		p.InactiveExitTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["ReloadPropagatedFrom"]; ok {
		p.ReloadPropagatedFrom = val.([]string)
	}
	if val, ok := m["SourcePath"]; ok {
		p.SourcePath = val.(string)
	}
	if val, ok := m["Id"]; ok {
		p.Id = val.(string)
	}
	if val, ok := m["JobTimeoutUSec"]; ok {
		p.JobTimeoutUSec = val.(uint64)
	}
	if val, ok := m["AssertResult"]; ok {
		p.AssertResult = val.(bool)
	}
	if val, ok := m["ActiveExitTimestamp"]; ok {
		p.ActiveExitTimestamp = val.(uint64)
	}
	if val, ok := m["RequisiteOf"]; ok {
		p.RequisiteOf = val.([]string)
	}
	if val, ok := m["StateChangeTimestampMonotonic"]; ok {
		p.StateChangeTimestampMonotonic = val.(uint64)
	}
	if val, ok := m["StateChangeTimestamp"]; ok {
		p.StateChangeTimestamp = val.(uint64)
	}
	if val, ok := m["PropagatesReloadTo"]; ok {
		p.PropagatesReloadTo = val.([]string)
	}
	if val, ok := m["Names"]; ok {
		p.Names = val.([]string)
	}
	if val, ok := m["NeedDaemonReload"]; ok {
		p.NeedDaemonReload = val.(bool)
	}
	if val, ok := m["AssertTimestamp"]; ok {
		p.AssertTimestamp = val.(uint64)
	}
	if val, ok := m["AllowIsolate"]; ok {
		p.AllowIsolate = val.(bool)
	}
	if val, ok := m["JobTimeoutRebootArgument"]; ok {
		p.JobTimeoutRebootArgument = val.(string)
	}
	if val, ok := m["UnitFileState"]; ok {
		p.UnitFileState = val.(string)
	}
	if val, ok := m["StopWhenUnneeded"]; ok {
		p.StopWhenUnneeded = val.(bool)
	}
	if val, ok := m["JoinsNamespaceOf"]; ok {
		p.JoinsNamespaceOf = val.([]string)
	}
	if val, ok := m["SubState"]; ok {
		p.SubState = val.(string)
	}
	if val, ok := m["Following"]; ok {
		p.Following = val.(string)
	}
	return p
}
