# Converge Changelog

## 0.3.0-rc1 24-10-2016

### Enhancements

- "param is required" error should include param name [#335](https://github.com/asteris-llc/converge/issues/335)
- file/dir refactor [#327](https://github.com/asteris-llc/converge/issues/327)
- Use github.com/pkg/errors exclusively [#300](https://github.com/asteris-llc/converge/issues/300)
- Use errgroup in server [#295](https://github.com/asteris-llc/converge/issues/295)
- Add ability to modify group [#279](https://github.com/asteris-llc/converge/issues/279)
- Allow ability to indicate group name when adding a user [#276](https://github.com/asteris-llc/converge/issues/276)
- add conditionals to module resource [#268](https://github.com/asteris-llc/converge/issues/268)
- named locks [#249](https://github.com/asteris-llc/converge/issues/249)
- pretty printed changes should align values [#244](https://github.com/asteris-llc/converge/issues/244)
- ability to wait for a condition [#238](https://github.com/asteris-llc/converge/issues/238)
- proper arrays [#110](https://github.com/asteris-llc/converge/issues/110)
- builtin file group module [#75](https://github.com/asteris-llc/converge/issues/75)
- named groups [#392](https://github.com/asteris-llc/converge/pull/392) ([ryane](https://github.com/ryane))
- Don't render module dependencies [#387](https://github.com/asteris-llc/converge/pull/387) ([BrianHicks](https://github.com/BrianHicks))
- Added docs, renamed packages to package for consistency with other modules [#384](https://github.com/asteris-llc/converge/pull/384) ([rebeccaskinner](https://github.com/rebeccaskinner))
- Use status.RaiseLevel in group [#377](https://github.com/asteris-llc/converge/pull/377) ([arichardet](https://github.com/arichardet))
- Feature/rpm module [#373](https://github.com/asteris-llc/converge/pull/373) ([rebeccaskinner](https://github.com/rebeccaskinner))
- use a container for CI [#372](https://github.com/asteris-llc/converge/pull/372) ([BrianHicks](https://github.com/BrianHicks))
- create a metadata envelope for nodes [#369](https://github.com/asteris-llc/converge/pull/369) ([BrianHicks](https://github.com/BrianHicks))
- cmd/server.go: use errgroup instead of waitgroup [#363](https://github.com/asteris-llc/converge/pull/363) ([QuentinPerez](https://github.com/QuentinPerez))
- Feature/conditionals [#362](https://github.com/asteris-llc/converge/pull/362) ([rebeccaskinner](https://github.com/rebeccaskinner))
- Feature/example swarm wait [#346](https://github.com/asteris-llc/converge/pull/346) ([ryane](https://github.com/ryane))
- Add Compound Parameters [#340](https://github.com/asteris-llc/converge/pull/340) ([BrianHicks](https://github.com/BrianHicks))
- load overlay module in elk example [#326](https://github.com/asteris-llc/converge/pull/326) ([feniix](https://github.com/feniix))
- Refactor deserializers [#321](https://github.com/asteris-llc/converge/pull/321) ([BrianHicks](https://github.com/BrianHicks))
- Use text/tabwriter to align human output [#317](https://github.com/asteris-llc/converge/pull/317) ([sehqlr](https://github.com/sehqlr))
- Fix/225 pipeline function refactor [#307](https://github.com/asteris-llc/converge/pull/307) ([rebeccaskinner](https://github.com/rebeccaskinner))

### Bug Fixes

- fix panic [#408](https://github.com/asteris-llc/converge/pull/408) ([rebeccaskinner](https://github.com/rebeccaskinner))
- Conditional Regression [#401](https://github.com/asteris-llc/converge/issues/401)
- Apply doesn't show diff output [#399](https://github.com/asteris-llc/converge/issues/399)
- Use lists as params in modules is broken [#397](https://github.com/asteris-llc/converge/issues/397)
- module dependencies now failing [#395](https://github.com/asteris-llc/converge/issues/395)
- Explicit dependencies fail inside of case statements [#385](https://github.com/asteris-llc/converge/issues/385)
- Use `package.rpm` not `rpm.package` for rpm module [#382](https://github.com/asteris-llc/converge/issues/382)
- docker.container regression [#343](https://github.com/asteris-llc/converge/issues/343)
- handle parameters with valid zero value [#338](https://github.com/asteris-llc/converge/issues/338)
- shell module should not set StatusLevel based on process exit code [#323](https://github.com/asteris-llc/converge/issues/323)
- StatusLevel not taken into account during graph execution [#322](https://github.com/asteris-llc/converge/issues/322)
- param dependency fails in `samples/shellContext.hcl` [#313](https://github.com/asteris-llc/converge/issues/313)
- Execution Engine Ignoring Warning Levels [#243](https://github.com/asteris-llc/converge/issues/243)
- Make pipeline functions use mult-return instead of Either [#225](https://github.com/asteris-llc/converge/issues/225)
- samples in the README are formatted incorrectly. [#104](https://github.com/asteris-llc/converge/issues/104)
- Handle thunked branches [#403](https://github.com/asteris-llc/converge/pull/403) ([rebeccaskinner](https://github.com/rebeccaskinner))
- don't exclude modules in getNearestAncestor [#396](https://github.com/asteris-llc/converge/pull/396) ([ryane](https://github.com/ryane))
- Fix/385 dependencies in conditionals [#391](https://github.com/asteris-llc/converge/pull/391) ([rebeccaskinner](https://github.com/rebeccaskinner))
- Fix lookup calls to use os/user in SetAddUserOptions [#390](https://github.com/asteris-llc/converge/pull/390) ([arichardet](https://github.com/arichardet))
- Change `rpm.package` to `package.rpm` [#383](https://github.com/asteris-llc/converge/pull/383) ([rebeccaskinner](https://github.com/rebeccaskinner))
- Update user status level and errors, add checks in Apply for group [#375](https://github.com/asteris-llc/converge/pull/375) ([arichardet](https://github.com/arichardet))
- Ability to use pointers as a preparer value [#339](https://github.com/asteris-llc/converge/pull/339) ([BrianHicks](https://github.com/BrianHicks))
- Status Error Codes [#333](https://github.com/asteris-llc/converge/pull/333) ([BrianHicks](https://github.com/BrianHicks))
- fix codeclimate yaml [#328](https://github.com/asteris-llc/converge/pull/328) ([BrianHicks](https://github.com/BrianHicks))


### Enhancements

- Ability to wait for a condition (#238)

### Bug fixes

- Fix #225 pipeline function refactor (#307)
- Map keys are considered "strings" in parse.Node (#315)

### Documentation/Examples

## 0.2.0 (September 26, 2016)

### Enhancements

- shell env and working dir support (#185)
- Docker image resource (#188)
- RPC support (#187)
- Expose basic platform information (#194)
- Docker container resource (#203)
- Support for create/delete linux groups (#234)
- Refactor Status Interface (#237)
- Refactor `check` and `apply` behavior (#240)
- Feature/module verification (#245)
- Reduce logging verbosity (#248)
- Boolean support in parameters (#251)
- Change Port to 4774 after IANA approval (#261)
- Support for create/delete linux users (#259)

### Bug fixes

- don't panic when result status is nil (#180)
- Fix error handling when fail to print results during plan or apply (#183)
- Order fixes (#254)
- Race condition fixes (#266)
- Fix/user group - Allow adding/deleting without gid (#283)
- Perform healthchecks over RPC (#289)
- Fix/294 field name conflicts (#301)
- Use thread-safe field cache (#303)

### Documentation/Examples

- Documentation Site (#192)
- Docker Swarm mode (#267)
- ELK (Elasticsearch, Logstash, and Kibana) stack (#272)
- Add CodeClimate Checks to build (#281)
- docs: add draft resource authors guide (#290)
