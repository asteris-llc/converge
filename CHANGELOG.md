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

### Bug Fixes 
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