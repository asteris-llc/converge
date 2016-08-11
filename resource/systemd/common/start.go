package common

type StartMode string
type StartModes []StartMode

const (
	SMReplace            StartMode = "replace"
	SMFail               StartMode = "fail"
	SMIsolate            StartMode = "isolate"
	SMIgnoreDependencies StartMode = "ignore-dependencies"
	SMIgnoreRequirements StartMode = "ignore-requirements"
)

var ValidStartModes = StartModes{SMReplace, SMFail, SMIsolate, SMIgnoreDependencies, SMIgnoreRequirements}
